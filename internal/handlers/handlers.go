// Package handlers implements the tool handler functions for go-study-mcp.
//
// Each handler receives decoded, schema-valid input and returns a typed
// tool.Result that splits model-facing text from UI-facing structured data.
//
// Audio generation is asynchronous: a non-preview generate/synthesize call
// enqueues a job (Registry), spawns a worker goroutine, and returns a job
// handle immediately so the host never blocks on a long LLM + TTS run. The UI
// polls list_jobs and plays completed audio via read_audio.
//
// Output paths are owned by the server (a writable OUTPUT_DIR resolved at
// startup), never by the caller — a caller-chosen path is unreliable across
// hosts (read-only CWD, non-existent container mounts).
package handlers

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	bifrost "github.com/maximhq/bifrost/core"
	"github.com/maximhq/bifrost/core/schemas"

	"github.com/hurtener/dockyard/runtime/tool"

	"github.com/hurtener/go-study-mcp/internal/audio"
	"github.com/hurtener/go-study-mcp/internal/contracts"
	"github.com/hurtener/go-study-mcp/internal/jobs"
	"github.com/hurtener/go-study-mcp/internal/prompts"
)

// bifrostClient is the shared Bifrost instance.
var bifrostClient *bifrost.Bifrost

// SkipLLM is a test flag that skips actual LLM calls.
var SkipLLM bool

// Registry holds the asynchronous generation jobs surfaced via list_jobs.
var Registry = jobs.NewRegistry()

// outputDir is the writable directory the server owns for generated audio.
// Resolved once at startup; never derived from caller-supplied paths.
var outputDir string

// maxReadBytes caps a read_audio inline payload (~25 MiB). Larger files report
// Truncated so the UI falls back to the saved path rather than shipping a huge
// base64 blob through the MCP pipe.
const maxReadBytes = 25 << 20

func init() {
	outputDir = resolveOutputDir()

	// Check if we have any API key before trying to init Bifrost
	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		apiKey = os.Getenv("OPENAI_API_KEY")
	}
	if apiKey == "" || apiKey == "sk-test" || apiKey == "sk-placeholder" {
		SkipLLM = true
		return
	}

	client, err := bifrost.Init(context.Background(), schemas.BifrostConfig{
		Account: &OpenRouterAccount{},
	})
	if err != nil {
		SkipLLM = true
		return
	}
	bifrostClient = client
}

// OpenRouterAccount implements the Bifrost Account interface for OpenRouter.
type OpenRouterAccount struct{}

func (a *OpenRouterAccount) GetConfiguredProviders() ([]schemas.ModelProvider, error) {
	return []schemas.ModelProvider{schemas.OpenAI}, nil
}

func (a *OpenRouterAccount) GetKeysForProvider(ctx context.Context, provider schemas.ModelProvider) ([]schemas.Key, error) {
	if provider == schemas.OpenAI {
		apiKey := os.Getenv("OPENROUTER_API_KEY")
		if apiKey == "" {
			apiKey = os.Getenv("OPENAI_API_KEY")
		}
		if apiKey == "" {
			SkipLLM = true
			return nil, fmt.Errorf("no API key found: set OPENROUTER_API_KEY or OPENAI_API_KEY")
		}
		return []schemas.Key{{
			Value:  *schemas.NewEnvVar(apiKey),
			Models: schemas.WhiteList{"*"},
			Weight: 1.0,
		}}, nil
	}
	return nil, fmt.Errorf("provider %s not supported", provider)
}

func (a *OpenRouterAccount) GetConfigForProvider(provider schemas.ModelProvider) (*schemas.ProviderConfig, error) {
	if provider == schemas.OpenAI {
		baseURL := os.Getenv("OPENROUTER_BASE_URL")
		if baseURL == "" {
			// Bifrost appends /v1/chat/completions internally
			baseURL = "https://openrouter.ai/api"
		}
		return &schemas.ProviderConfig{
			NetworkConfig: schemas.NetworkConfig{
				BaseURL:                        baseURL,
				DefaultRequestTimeoutInSeconds: 120,
			},
			ConcurrencyAndBufferSize: schemas.DefaultConcurrencyAndBufferSize,
		}, nil
	}
	return nil, fmt.Errorf("provider %s not supported", provider)
}

// ──────────────────────────────────────────────────────────────────────────────
// generate_podcast
// ──────────────────────────────────────────────────────────────────────────────

// GeneratePodcast is the handler for the generate_podcast tool.
func GeneratePodcast(_ context.Context, in contracts.GeneratePodcastInput) (tool.Result[contracts.GeneratePodcastOutput], error) {
	lang := normalizeLang(in.Language)

	if in.PreviewOnly {
		script, err := podcastScript(in, lang)
		if err != nil {
			return tool.Result[contracts.GeneratePodcastOutput]{}, err
		}
		wordCount := len(strings.Fields(script))
		return tool.Result[contracts.GeneratePodcastOutput]{
			Text: script,
			Structured: contracts.GeneratePodcastOutput{
				Kind:             "podcast",
				Script:           script,
				WordCount:        wordCount,
				DurationEstimate: estimateDuration(wordCount),
				Language:         lang,
				PreviewOnly:      true,
			},
		}, nil
	}

	if notConfigured() {
		return ttsNotConfigured[contracts.GeneratePodcastOutput]("podcast", contracts.GeneratePodcastOutput{
			Kind: "podcast", Status: string(jobs.StatusFailed), Language: lang,
		})
	}

	job := Registry.Add("podcast", titleFor(in.TopicHint, in.Content, "Podcast"))
	go func() {
		Registry.SetProcessing(job.ID)
		script, err := podcastScript(in, lang)
		if err != nil {
			Registry.Fail(job.ID, err.Error())
			return
		}
		wordCount := len(strings.Fields(script))
		runSynthesisJob(job.ID, "podcast", script, in.Voice, "", estimateDuration(wordCount), 0, wordCount)
	}()

	return jobStarted("podcast", job.ID, contracts.GeneratePodcastOutput{
		Kind: "podcast", JobID: job.ID, Status: string(jobs.StatusProcessing), Language: lang,
	})
}

func podcastScript(in contracts.GeneratePodcastInput, lang string) (string, error) {
	system, user := prompts.PodcastScript(prompts.PodcastParams{
		Language:       lang,
		Tone:           in.Tone,
		Persona:        in.Persona,
		TopicHint:      in.TopicHint,
		DurationTarget: in.DurationTarget,
		WordTarget:     wordTargetForDuration(in.DurationTarget),
		Content:        in.Content,
	})
	return chatComplete(system, user)
}

// ──────────────────────────────────────────────────────────────────────────────
// generate_flashcards
// ──────────────────────────────────────────────────────────────────────────────

// GenerateFlashcards is the handler for the generate_flashcards tool.
func GenerateFlashcards(_ context.Context, in contracts.GenerateFlashcardsInput) (tool.Result[contracts.GenerateFlashcardsOutput], error) {
	lang := normalizeLang(in.Language)

	cards, err := flashcards(in, lang)
	if err != nil {
		return tool.Result[contracts.GenerateFlashcardsOutput]{}, err
	}

	cardCount := len(cards)
	pauseDuration := in.PauseDuration
	if pauseDuration == 0 {
		pauseDuration = 5
	}
	estimate := fmt.Sprintf("~%d min", (cardCount*(pauseDuration+3))/60)

	if in.PreviewOnly {
		return tool.Result[contracts.GenerateFlashcardsOutput]{
			Text: formatCardsText(cards),
			Structured: contracts.GenerateFlashcardsOutput{
				Kind:             "flashcards",
				Cards:            cards,
				CardCount:        cardCount,
				DurationEstimate: estimate,
				Language:         lang,
				PreviewOnly:      true,
			},
		}, nil
	}

	if notConfigured() {
		return ttsNotConfigured[contracts.GenerateFlashcardsOutput]("flashcards", contracts.GenerateFlashcardsOutput{
			Kind: "flashcards", Cards: cards, CardCount: cardCount, Status: string(jobs.StatusFailed), Language: lang,
		})
	}

	script := buildFlashcardScript(cards, pauseDuration)
	job := Registry.Add("flashcards", fmt.Sprintf("%d flashcards", cardCount))
	go func() {
		Registry.SetProcessing(job.ID)
		runSynthesisJob(job.ID, "flashcards", script, in.Voice, "", estimate, 0, 0)
	}()

	return jobStarted("flashcards", job.ID, contracts.GenerateFlashcardsOutput{
		Kind: "flashcards", Cards: cards, CardCount: cardCount,
		JobID: job.ID, Status: string(jobs.StatusProcessing), Language: lang,
	})
}

// flashcards returns the cards to narrate: pre-curated input, LLM-extracted
// from content, or an error when neither is available.
func flashcards(in contracts.GenerateFlashcardsInput, lang string) ([]contracts.FlashCard, error) {
	if len(in.Cards) > 0 {
		return in.Cards, nil
	}
	if in.Content == "" {
		return nil, fmt.Errorf("either content or cards must be provided")
	}

	system, user := prompts.FlashcardExtract(prompts.FlashcardParams{
		Language:   lang,
		Difficulty: in.Difficulty,
		CardCount:  in.CardCount,
		Content:    in.Content,
	})

	if SkipLLM || bifrostClient == nil {
		return []contracts.FlashCard{}, nil
	}

	raw, err := chatComplete(system, user)
	if err != nil {
		return nil, err
	}
	raw = strings.TrimPrefix(raw, "```json")
	raw = strings.TrimPrefix(raw, "```")
	raw = strings.TrimSuffix(raw, "```")
	raw = strings.TrimSpace(raw)

	var extracted []contracts.FlashCard
	if err := json.Unmarshal([]byte(raw), &extracted); err != nil {
		return nil, fmt.Errorf("failed to parse flashcards: %w", err)
	}
	return extracted, nil
}

// ──────────────────────────────────────────────────────────────────────────────
// synthesize_speech
// ──────────────────────────────────────────────────────────────────────────────

// SynthesizeSpeech is the handler for the synthesize_speech tool.
func SynthesizeSpeech(_ context.Context, in contracts.SynthesizeSpeechInput) (tool.Result[contracts.SynthesizeSpeechOutput], error) {
	charCount := len(removePauseMarkers(in.Text))
	estimate := fmt.Sprintf("~%d sec", charCount/20)

	if notConfigured() {
		return tool.Result[contracts.SynthesizeSpeechOutput]{
			Text: fmt.Sprintf("TTS is not configured (set OPENROUTER_API_KEY or OPENAI_API_KEY). Would synthesize %d characters.", charCount),
			Structured: contracts.SynthesizeSpeechOutput{
				Kind:             "synthesize",
				Status:           string(jobs.StatusFailed),
				CharacterCount:   charCount,
				DurationEstimate: estimate,
			},
		}, nil
	}

	job := Registry.Add("synthesize", snippet(in.Text))
	go func() {
		Registry.SetProcessing(job.ID)
		runSynthesisJob(job.ID, "synthesize", in.Text, in.Voice, in.ResponseFormat, estimate, charCount, 0)
	}()

	return tool.Result[contracts.SynthesizeSpeechOutput]{
		Text: fmt.Sprintf("Started speech synthesis (job %s). It will appear in the Jobs tab when ready.", job.ID),
		Structured: contracts.SynthesizeSpeechOutput{
			Kind:             "synthesize",
			JobID:            job.ID,
			Status:           string(jobs.StatusProcessing),
			CharacterCount:   charCount,
			DurationEstimate: estimate,
		},
	}, nil
}

// ──────────────────────────────────────────────────────────────────────────────
// generate_study_guide
// ──────────────────────────────────────────────────────────────────────────────

// GenerateStudyGuide is the handler for the generate_study_guide tool.
func GenerateStudyGuide(_ context.Context, in contracts.GenerateStudyGuideInput) (tool.Result[contracts.GenerateStudyGuideOutput], error) {
	lang := normalizeLang(in.Language)
	difficulty := in.Difficulty
	if difficulty == "" {
		difficulty = "graduate"
	}

	if in.PreviewOnly {
		script, err := studyGuideScript(in, lang, difficulty)
		if err != nil {
			return tool.Result[contracts.GenerateStudyGuideOutput]{}, err
		}
		wordCount := len(strings.Fields(script))
		return tool.Result[contracts.GenerateStudyGuideOutput]{
			Text: script,
			Structured: contracts.GenerateStudyGuideOutput{
				Kind:             "study_guide",
				Script:           script,
				Sections:         extractSections(script),
				WordCount:        wordCount,
				DurationEstimate: estimateStudyGuideDuration(wordCount),
				Language:         lang,
				Difficulty:       difficulty,
				PreviewOnly:      true,
			},
		}, nil
	}

	if notConfigured() {
		return ttsNotConfigured[contracts.GenerateStudyGuideOutput]("study_guide", contracts.GenerateStudyGuideOutput{
			Kind: "study_guide", Status: string(jobs.StatusFailed), Language: lang, Difficulty: difficulty,
		})
	}

	job := Registry.Add("study_guide", titleFor("", in.Content, "Study Guide"))
	go func() {
		Registry.SetProcessing(job.ID)
		script, err := studyGuideScript(in, lang, difficulty)
		if err != nil {
			Registry.Fail(job.ID, err.Error())
			return
		}
		wordCount := len(strings.Fields(script))
		runSynthesisJob(job.ID, "study_guide", script, in.Voice, "", estimateStudyGuideDuration(wordCount), 0, wordCount)
	}()

	return jobStarted("study_guide", job.ID, contracts.GenerateStudyGuideOutput{
		Kind: "study_guide", JobID: job.ID, Status: string(jobs.StatusProcessing),
		Language: lang, Difficulty: difficulty,
	})
}

func studyGuideScript(in contracts.GenerateStudyGuideInput, lang, difficulty string) (string, error) {
	system, user := prompts.StudyGuideScript(prompts.StudyGuideParams{
		Language:       lang,
		Difficulty:     difficulty,
		DurationTarget: in.DurationTarget,
		WordTarget:     studyGuideWordTarget(in.DurationTarget),
		Content:        in.Content,
	})
	return chatComplete(system, user)
}

// ──────────────────────────────────────────────────────────────────────────────
// list_jobs
// ──────────────────────────────────────────────────────────────────────────────

// ListJobs is the handler for the list_jobs tool. The UI polls it to render
// the Jobs tab.
func ListJobs(_ context.Context, _ contracts.ListJobsInput) (tool.Result[contracts.ListJobsOutput], error) {
	registryJobs := Registry.List()
	out := make([]contracts.Job, 0, len(registryJobs))
	for _, j := range registryJobs {
		out = append(out, contracts.Job{
			ID:               j.ID,
			Kind:             j.Kind,
			Title:            j.Title,
			Status:           string(j.Status),
			CreatedAt:        j.CreatedAt.Format(time.RFC3339),
			UpdatedAt:        j.UpdatedAt.Format(time.RFC3339),
			OutputPath:       j.OutputPath,
			DurationEstimate: j.DurationEstimate,
			CharacterCount:   j.CharacterCount,
			WordCount:        j.WordCount,
			Error:            j.Error,
		})
	}
	return tool.Result[contracts.ListJobsOutput]{
		Text:       fmt.Sprintf("%d job(s)", len(out)),
		Structured: contracts.ListJobsOutput{Kind: "jobs", Jobs: out},
	}, nil
}

// ──────────────────────────────────────────────────────────────────────────────
// read_audio
// ──────────────────────────────────────────────────────────────────────────────

// ReadAudio is the handler for the read_audio tool: it returns a generated
// audio file as a data URI so a sandboxed UI can play it inline. Reads are
// confined to OUTPUT_DIR.
func ReadAudio(_ context.Context, in contracts.ReadAudioInput) (tool.Result[contracts.ReadAudioOutput], error) {
	fail := func(err error) (tool.Result[contracts.ReadAudioOutput], error) {
		return tool.Result[contracts.ReadAudioOutput]{}, fmt.Errorf("read_audio: %w", err)
	}

	path := in.Path
	if in.JobID != "" {
		j, ok := Registry.Get(in.JobID)
		if !ok {
			return fail(fmt.Errorf("unknown job %q", in.JobID))
		}
		if j.Status != jobs.StatusCompleted || j.OutputPath == "" {
			return fail(fmt.Errorf("job %q is %s, no audio yet", in.JobID, j.Status))
		}
		path = j.OutputPath
	}
	if strings.TrimSpace(path) == "" {
		return fail(fmt.Errorf("provide jobId or path"))
	}

	abs, err := filepath.Abs(path)
	if err != nil {
		return fail(err)
	}
	if !withinOutputDir(abs) {
		return fail(fmt.Errorf("path is outside the output directory"))
	}

	info, err := os.Stat(abs)
	if err != nil {
		return fail(err)
	}
	mime := mimeForAudio(abs)
	if info.Size() > maxReadBytes {
		return tool.Result[contracts.ReadAudioOutput]{
			Text: fmt.Sprintf("%s is %d bytes — too large to inline (cap %d); open the file at %s",
				filepath.Base(abs), info.Size(), maxReadBytes, abs),
			Structured: contracts.ReadAudioOutput{Kind: "audio", Mime: mime, SizeBytes: info.Size(), Truncated: true},
		}, nil
	}

	data, err := os.ReadFile(abs) //nolint:gosec // path validated + confined to OUTPUT_DIR
	if err != nil {
		return fail(err)
	}
	return tool.Result[contracts.ReadAudioOutput]{
		Text: fmt.Sprintf("Read %s (%s, %d bytes)", filepath.Base(abs), mime, info.Size()),
		Structured: contracts.ReadAudioOutput{
			Kind:      "audio",
			DataURI:   "data:" + mime + ";base64," + base64.StdEncoding.EncodeToString(data),
			Mime:      mime,
			SizeBytes: info.Size(),
		},
	}, nil
}

// ──────────────────────────────────────────────────────────────────────────────
// Job worker + result helpers
// ──────────────────────────────────────────────────────────────────────────────

// runSynthesisJob synthesizes text to audio, writes it under OUTPUT_DIR, and
// records the terminal job state. Runs inside the worker goroutine.
func runSynthesisJob(jobID, kind, text, voice, format, durationEstimate string, charCount, wordCount int) {
	data, ext, err := synthesize(text, voice, format)
	if err != nil {
		Registry.Fail(jobID, err.Error())
		return
	}
	path, err := writeOutput(kind, jobID, data, ext)
	if err != nil {
		Registry.Fail(jobID, err.Error())
		return
	}
	Registry.Complete(jobID, path, durationEstimate, charCount, wordCount)
}

// jobStarted builds the immediate "processing" tool result for an async tool.
func jobStarted[T any](kind, jobID string, structured T) (tool.Result[T], error) {
	return tool.Result[T]{
		Text: fmt.Sprintf("Started %s generation (job %s). It will appear in the Jobs tab when ready.",
			strings.ReplaceAll(kind, "_", " "), jobID),
		Structured: structured,
	}, nil
}

// ttsNotConfigured builds a non-error result explaining that no API key is set.
func ttsNotConfigured[T any](kind string, structured T) (tool.Result[T], error) {
	return tool.Result[T]{
		Text: fmt.Sprintf("Cannot generate %s audio: TTS is not configured (set OPENROUTER_API_KEY or OPENAI_API_KEY).",
			strings.ReplaceAll(kind, "_", " ")),
		Structured: structured,
	}, nil
}

// notConfigured reports whether the LLM/TTS client is unavailable.
func notConfigured() bool { return bifrostClient == nil || SkipLLM }

// ──────────────────────────────────────────────────────────────────────────────
// LLM + TTS
// ──────────────────────────────────────────────────────────────────────────────

// chatComplete runs a single chat completion, or returns a deterministic
// stand-in when no LLM is configured (so preview/test paths stay usable).
func chatComplete(system, user string) (string, error) {
	if SkipLLM || bifrostClient == nil {
		return fmt.Sprintf("[SYSTEM]\n%s\n\n[USER]\n%s", system, user), nil
	}
	response, bfErr := bifrostClient.ChatCompletionRequest(
		schemas.NewBifrostContext(context.Background(), schemas.NoDeadline),
		&schemas.BifrostChatRequest{
			Provider: schemas.OpenAI,
			Model:    chatModel(),
			Input: []schemas.ChatMessage{
				{
					Role:    schemas.ChatMessageRoleSystem,
					Content: &schemas.ChatMessageContent{ContentStr: schemas.Ptr(system)},
				},
				{
					Role:    schemas.ChatMessageRoleUser,
					Content: &schemas.ChatMessageContent{ContentStr: schemas.Ptr(user)},
				},
			},
		},
	)
	if bfErr != nil {
		return "", fmt.Errorf("llm call failed: %s", bfErr.GetErrorString())
	}
	if response != nil && len(response.Choices) > 0 && response.Choices[0].Message.Content.ContentStr != nil {
		return *response.Choices[0].Message.Content.ContentStr, nil
	}
	return "[No response from LLM]", nil
}

// synthesize converts text to encoded audio bytes, chunking the input to
// respect the TTS request-size limit (Gemini flash TTS caps input length).
// It returns the audio bytes and their file extension.
func synthesize(text, voice, format string) ([]byte, string, error) {
	if bifrostClient == nil {
		return nil, "", fmt.Errorf("tts client not initialized")
	}

	clean := removePauseMarkers(text)
	model := ttsModel()

	// Gemini TTS only supports PCM; everything else honors the requested
	// container (default mp3).
	responseFormat := format
	if responseFormat == "" {
		responseFormat = "mp3"
	}
	if strings.Contains(model, "gemini") {
		responseFormat = "pcm"
	}

	chunks := chunkText(clean, maxInputChars())
	var parts [][]byte
	for _, chunk := range chunks {
		if strings.TrimSpace(chunk) == "" {
			continue
		}
		response, bfErr := bifrostClient.SpeechRequest(
			schemas.NewBifrostContext(context.Background(), schemas.NoDeadline),
			&schemas.BifrostSpeechRequest{
				Provider: schemas.OpenAI,
				Model:    model,
				Input:    &schemas.SpeechInput{Input: chunk},
				Params: &schemas.SpeechParameters{
					VoiceConfig:    &schemas.SpeechVoiceInput{Voice: schemas.Ptr(voiceID(voice))},
					ResponseFormat: responseFormat,
				},
			},
		)
		if bfErr != nil {
			return nil, "", fmt.Errorf("tts failed: %s", bfErr.GetErrorString())
		}
		if response != nil && len(response.Audio) > 0 {
			parts = append(parts, response.Audio)
		}
	}

	joined := bytes.Join(parts, nil)
	if len(joined) == 0 {
		return nil, "", fmt.Errorf("no audio returned")
	}

	// PCM is concatenated raw, then encoded once to MP3.
	if responseFormat == "pcm" {
		mp3Data, err := audio.PCMToMP3(joined, defaultSampleRate())
		if err != nil {
			return joined, "raw", nil // fallback: keep raw PCM
		}
		return mp3Data, "mp3", nil
	}
	// mp3 frames concatenate cleanly for playback.
	ext := "mp3"
	if responseFormat == "wav" {
		ext = "wav"
	}
	return joined, ext, nil
}

// ──────────────────────────────────────────────────────────────────────────────
// Output directory (server-owned, writable)
// ──────────────────────────────────────────────────────────────────────────────

// resolveOutputDir picks a writable, absolute directory for generated audio.
// It tries STUDYAUDIO_OUTPUT_DIR (relative values anchored under $HOME, never
// the read-only CWD a host may launch us in), then ~/go-study-mcp, then the OS
// temp dir — returning the first that exists and is writable.
func resolveOutputDir() string {
	var candidates []string
	if d := strings.TrimSpace(os.Getenv("STUDYAUDIO_OUTPUT_DIR")); d != "" {
		candidates = append(candidates, anchorDir(expandHome(d)))
	}
	if home, err := os.UserHomeDir(); err == nil {
		candidates = append(candidates, filepath.Join(home, "go-study-mcp"))
	}
	candidates = append(candidates, filepath.Join(os.TempDir(), "go-study-mcp"))

	for _, c := range candidates {
		if c == "" {
			continue
		}
		if abs, err := filepath.Abs(c); err == nil {
			c = abs
		}
		if err := os.MkdirAll(c, 0o755); err != nil {
			continue
		}
		if dirWritable(c) {
			return c
		}
	}
	return os.TempDir()
}

// expandHome replaces a leading ~ with the user's home directory.
func expandHome(p string) string {
	if p == "~" || strings.HasPrefix(p, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			if p == "~" {
				return home
			}
			return filepath.Join(home, p[2:])
		}
	}
	return p
}

// anchorDir makes a relative path absolute by anchoring it under $HOME, so a
// configured "./output" never resolves against a read-only launch CWD.
func anchorDir(p string) string {
	if filepath.IsAbs(p) {
		return p
	}
	if home, err := os.UserHomeDir(); err == nil {
		return filepath.Join(home, p)
	}
	return p
}

// dirWritable reports whether dir accepts a probe write.
func dirWritable(dir string) bool {
	probe := filepath.Join(dir, ".write-probe")
	if err := os.WriteFile(probe, []byte("x"), 0o644); err != nil {
		return false
	}
	_ = os.Remove(probe)
	return true
}

// withinOutputDir reports whether abs is inside (or equal to) outputDir.
func withinOutputDir(abs string) bool {
	root, err := filepath.Abs(outputDir)
	if err != nil {
		return false
	}
	return abs == root || strings.HasPrefix(abs, root+string(os.PathSeparator))
}

// writeOutput writes audio bytes under OUTPUT_DIR with a server-chosen name.
func writeOutput(kind, jobID string, data []byte, ext string) (string, error) {
	name := fmt.Sprintf("%s-%s-%s.%s", time.Now().Format("20060102-150405"), kind, shortID(jobID), ext)
	path := filepath.Join(outputDir, name)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return "", fmt.Errorf("failed to write audio: %w", err)
	}
	return path, nil
}

// shortID returns the last dash-delimited segment of a job ID for filenames.
func shortID(jobID string) string {
	if i := strings.LastIndex(jobID, "-"); i >= 0 && i+1 < len(jobID) {
		return jobID[i+1:]
	}
	return jobID
}

func mimeForAudio(path string) string {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".mp3":
		return "audio/mpeg"
	case ".wav":
		return "audio/wav"
	case ".ogg":
		return "audio/ogg"
	case ".m4a":
		return "audio/mp4"
	default:
		return "application/octet-stream"
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// Text helpers
// ──────────────────────────────────────────────────────────────────────────────

func normalizeLang(lang string) string {
	lang = strings.TrimSpace(strings.ToLower(lang))
	if lang == "" {
		return "en"
	}
	if i := strings.Index(lang, "-"); i > 0 {
		lang = lang[:i]
	}
	if i := strings.Index(lang, "_"); i > 0 {
		lang = lang[:i]
	}
	return lang
}

func wordTargetForDuration(d string) int {
	switch d {
	case "short":
		return 300
	case "long":
		return 1400
	default:
		return 700
	}
}

func estimateDuration(words int) string {
	minutes := words / 150
	if minutes < 1 {
		return "~1 min"
	}
	return fmt.Sprintf("~%d min", minutes)
}

func studyGuideWordTarget(d string) int {
	switch d {
	case "short":
		return 2000
	case "long":
		return 8000
	default:
		return 4500
	}
}

func estimateStudyGuideDuration(words int) string {
	minutes := words / 130 // study guides are slower-paced
	if minutes < 1 {
		return "~1 min"
	}
	return fmt.Sprintf("~%d min", minutes)
}

func extractSections(script string) []contracts.StudyGuideSection {
	var sections []contracts.StudyGuideSection
	lines := strings.Split(script, "\n")
	var currentTitle string
	var currentContent strings.Builder

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		// Detect section headers (lines starting with # or [pause] followed by text)
		if strings.HasPrefix(trimmed, "# ") {
			if currentTitle != "" {
				sections = append(sections, contracts.StudyGuideSection{
					Title:   currentTitle,
					Content: strings.TrimSpace(currentContent.String()),
				})
				currentContent.Reset()
			}
			currentTitle = strings.TrimPrefix(trimmed, "# ")
		} else if currentTitle == "" && len(trimmed) > 0 {
			currentTitle = trimmed
		} else {
			if currentContent.Len() > 0 {
				currentContent.WriteString("\n")
			}
			currentContent.WriteString(line)
		}
	}

	if currentTitle != "" {
		sections = append(sections, contracts.StudyGuideSection{
			Title:   currentTitle,
			Content: strings.TrimSpace(currentContent.String()),
		})
	}

	// If no sections found, return whole script as one section
	if len(sections) == 0 {
		sections = []contracts.StudyGuideSection{
			{Title: "Study Guide", Content: script},
		}
	}

	return sections
}

func formatCardsText(cards []contracts.FlashCard) string {
	var sb strings.Builder
	for i, c := range cards {
		sb.WriteString(fmt.Sprintf("[%d] Q: %s\n    A: %s\n\n", i+1, c.Question, c.Answer))
	}
	return sb.String()
}

func buildFlashcardScript(cards []contracts.FlashCard, pauseSec int) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Welcome to your flashcard session. You will hear %d questions.\n", len(cards)))
	sb.WriteString(fmt.Sprintf("After each question, you will have %d seconds to recall the answer before it is revealed.\n\n", pauseSec))

	for i, c := range cards {
		sb.WriteString(fmt.Sprintf("[Card %d of %d]\n", i+1, len(cards)))
		sb.WriteString(fmt.Sprintf("Question: %s\n", c.Question))
		sb.WriteString(fmt.Sprintf("[PAUSE:%d]\n", pauseSec))
		sb.WriteString(fmt.Sprintf("Answer: %s\n\n", c.Answer))
	}

	sb.WriteString("That's the end of your session. Great work.\n")
	return sb.String()
}

func removePauseMarkers(text string) string {
	result := text
	for {
		idx := strings.Index(result, "[PAUSE:")
		if idx < 0 {
			break
		}
		end := strings.Index(result[idx:], "]")
		if end < 0 {
			break
		}
		result = result[:idx] + result[idx+end+1:]
	}
	return result
}

// chunkText splits text into chunks of at most max runes, preferring to break
// at sentence boundaries, then at whitespace, and only hard-cutting as a last
// resort. A non-positive max (or short text) yields a single chunk.
func chunkText(text string, max int) []string {
	r := []rune(strings.TrimSpace(text))
	if max <= 0 || len(r) <= max {
		return []string{string(r)}
	}

	var chunks []string
	floor := max * 3 / 5 // don't break too early
	for len(r) > 0 {
		if len(r) <= max {
			if c := strings.TrimSpace(string(r)); c != "" {
				chunks = append(chunks, c)
			}
			break
		}
		best := -1
		for i := max; i > floor; i-- {
			switch r[i-1] {
			case '.', '!', '?', '\n':
				best = i
			}
			if best >= 0 {
				break
			}
		}
		if best < 0 {
			for i := max; i > floor; i-- {
				if r[i-1] == ' ' {
					best = i
					break
				}
			}
		}
		if best < 0 {
			best = max
		}
		if c := strings.TrimSpace(string(r[:best])); c != "" {
			chunks = append(chunks, c)
		}
		r = r[best:]
	}
	return chunks
}

func snippet(s string) string {
	s = strings.TrimSpace(strings.ReplaceAll(s, "\n", " "))
	r := []rune(s)
	if len(r) > 60 {
		return string(r[:60]) + "…"
	}
	return s
}

func titleFor(hint, content, fallback string) string {
	if t := strings.TrimSpace(hint); t != "" {
		return t
	}
	if t := snippet(content); t != "" {
		return t
	}
	return fallback
}

// ──────────────────────────────────────────────────────────────────────────────
// Configuration
// ──────────────────────────────────────────────────────────────────────────────

func chatModel() string {
	// Support STUDYAUDIO_LLM_MODEL from .env
	if model := os.Getenv("STUDYAUDIO_LLM_MODEL"); model != "" {
		return model
	}
	if model := os.Getenv("OPENROUTER_MODEL"); model != "" {
		return model
	}
	if model := os.Getenv("OPENAI_MODEL"); model != "" {
		return model
	}
	return "deepseek/deepseek-v4-pro"
}

func ttsModel() string {
	if model := os.Getenv("STUDYAUDIO_TTS_MODEL"); model != "" {
		return model
	}
	if model := os.Getenv("TTS_MODEL"); model != "" {
		return model
	}
	return "google/gemini-3.1-flash-tts-preview"
}

func defaultSampleRate() int {
	if rate := os.Getenv("STUDYAUDIO_SAMPLE_RATE"); rate != "" {
		var r int
		if _, err := fmt.Sscanf(rate, "%d", &r); err == nil && r > 0 {
			return r
		}
	}
	return 24000
}

// maxInputChars is the per-request TTS input cap (Gemini flash TTS limit).
func maxInputChars() int {
	if v := os.Getenv("STUDYAUDIO_MAX_INPUT_CHARS"); v != "" {
		var n int
		if _, err := fmt.Sscanf(v, "%d", &n); err == nil && n > 0 {
			return n
		}
	}
	return 1500
}

func voiceID(voice string) string {
	if voice != "" {
		return voice
	}
	if v := os.Getenv("STUDYAUDIO_DEFAULT_VOICE"); v != "" {
		return v
	}
	if v := os.Getenv("TTS_VOICE"); v != "" {
		return v
	}
	return "Erinome"
}
