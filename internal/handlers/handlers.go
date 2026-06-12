// Package handlers implements the tool handler functions for go-study-mcp.
//
// Each handler receives decoded, schema-valid input and returns a typed
// tool.Result that splits model-facing text from UI-facing structured data.
package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	bifrost "github.com/maximhq/bifrost/core"
	"github.com/maximhq/bifrost/core/schemas"

	"github.com/hurtener/dockyard/runtime/tool"

	"github.com/hurtener/go-study-mcp/internal/contracts"
	"github.com/hurtener/go-study-mcp/internal/prompts"
)

// bifrostClient is the shared Bifrost instance.
var bifrostClient *bifrost.Bifrost

// SkipLLM is a test flag that skips actual LLM calls.
var SkipLLM bool

func init() {
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
			Value:   *schemas.NewEnvVar(apiKey),
			Models:  schemas.WhiteList{"*"},
			Weight:  1.0,
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
				BaseURL: baseURL,
			},
			ConcurrencyAndBufferSize: schemas.DefaultConcurrencyAndBufferSize,
		}, nil
	}
	return nil, fmt.Errorf("provider %s not supported", provider)
}

// GeneratePodcast is the handler for the generate_podcast tool.
func GeneratePodcast(_ context.Context, in contracts.GeneratePodcastInput) (tool.Result[contracts.GeneratePodcastOutput], error) {
	lang := normalizeLang(in.Language)
	wordTarget := wordTargetForDuration(in.DurationTarget)

	system, user := prompts.PodcastScript(prompts.PodcastParams{
		Language:       lang,
		Tone:           in.Tone,
		Persona:        in.Persona,
		TopicHint:      in.TopicHint,
		DurationTarget: in.DurationTarget,
		WordTarget:     wordTarget,
		Content:        in.Content,
	})

	var script string
	if SkipLLM || bifrostClient == nil {
		script = fmt.Sprintf("[SYSTEM]\n%s\n\n[USER]\n%s", system, user)
	} else {
		response, bfErr := bifrostClient.ChatCompletionRequest(
			schemas.NewBifrostContext(context.Background(), schemas.NoDeadline),
			&schemas.BifrostChatRequest{
				Provider: schemas.OpenAI,
				Model:    chatModel(),
				Input: []schemas.ChatMessage{
					{
						Role: schemas.ChatMessageRoleSystem,
						Content: &schemas.ChatMessageContent{
							ContentStr: schemas.Ptr(system),
						},
					},
					{
						Role: schemas.ChatMessageRoleUser,
						Content: &schemas.ChatMessageContent{
							ContentStr: schemas.Ptr(user),
						},
					},
				},
			},
		)
		if bfErr != nil {
			return tool.Result[contracts.GeneratePodcastOutput]{}, fmt.Errorf("llm call failed: %s", bfErr.GetErrorString())
		}
		if response != nil && response.Choices != nil && len(response.Choices) > 0 && response.Choices[0].Message.Content.ContentStr != nil {
			script = *response.Choices[0].Message.Content.ContentStr
		} else {
			script = "[No response from LLM]"
		}
	}

	wordCount := len(strings.Fields(script))

	if in.PreviewOnly {
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

	audioPath, err := synthesizeToDisk(script, in.Voice, "podcast")
	if err != nil {
		return tool.Result[contracts.GeneratePodcastOutput]{}, fmt.Errorf("tts failed: %w", err)
	}

	return tool.Result[contracts.GeneratePodcastOutput]{
		Text: script,
		Structured: contracts.GeneratePodcastOutput{
			Kind:             "podcast",
			Script:           script,
			OutputPath:       audioPath,
			WordCount:        wordCount,
			DurationEstimate: estimateDuration(wordCount),
			Language:         lang,
			PreviewOnly:      false,
		},
	}, nil
}

// GenerateFlashcards is the handler for the generate_flashcards tool.
func GenerateFlashcards(_ context.Context, in contracts.GenerateFlashcardsInput) (tool.Result[contracts.GenerateFlashcardsOutput], error) {
	lang := normalizeLang(in.Language)
	if lang == "" {
		lang = "en"
	}

	var cards []contracts.FlashCard

	if len(in.Cards) > 0 {
		cards = in.Cards
	} else if in.Content != "" {
		system, user := prompts.FlashcardExtract(prompts.FlashcardParams{
			Language:   lang,
			Difficulty: in.Difficulty,
			CardCount:  in.CardCount,
			Content:    in.Content,
		})

		if SkipLLM || bifrostClient == nil {
			cards = []contracts.FlashCard{}
		} else {
			response, bfErr := bifrostClient.ChatCompletionRequest(
				schemas.NewBifrostContext(context.Background(), schemas.NoDeadline),
				&schemas.BifrostChatRequest{
					Provider: schemas.OpenAI,
					Model:    chatModel(),
					Input: []schemas.ChatMessage{
						{
							Role: schemas.ChatMessageRoleSystem,
							Content: &schemas.ChatMessageContent{
								ContentStr: schemas.Ptr(system),
							},
						},
						{
							Role: schemas.ChatMessageRoleUser,
							Content: &schemas.ChatMessageContent{
								ContentStr: schemas.Ptr(user),
							},
						},
					},
				},
			)
			if bfErr != nil {
				return tool.Result[contracts.GenerateFlashcardsOutput]{}, fmt.Errorf("llm call failed: %s", bfErr.GetErrorString())
			}

			raw := ""
			if response != nil && response.Choices != nil && len(response.Choices) > 0 && response.Choices[0].Message.Content.ContentStr != nil {
				raw = *response.Choices[0].Message.Content.ContentStr
			}
			raw = strings.TrimPrefix(raw, "```json")
			raw = strings.TrimPrefix(raw, "```")
			raw = strings.TrimSuffix(raw, "```")
			raw = strings.TrimSpace(raw)

			var extracted []contracts.FlashCard
			if err := json.Unmarshal([]byte(raw), &extracted); err != nil {
				return tool.Result[contracts.GenerateFlashcardsOutput]{}, fmt.Errorf("failed to parse flashcards: %w", err)
			}
			cards = extracted
		}
	} else {
		return tool.Result[contracts.GenerateFlashcardsOutput]{}, fmt.Errorf("either content or cards must be provided")
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

	script := buildFlashcardScript(cards, pauseDuration)
	audioPath, err := synthesizeToDisk(script, in.Voice, "flashcards")
	if err != nil {
		return tool.Result[contracts.GenerateFlashcardsOutput]{}, fmt.Errorf("tts failed: %w", err)
	}

	return tool.Result[contracts.GenerateFlashcardsOutput]{
		Text: formatCardsText(cards),
		Structured: contracts.GenerateFlashcardsOutput{
			Kind:             "flashcards",
			Cards:            cards,
			OutputPath:       audioPath,
			CardCount:        cardCount,
			DurationEstimate: estimate,
			Language:         lang,
			PreviewOnly:      false,
		},
	}, nil
}

// SynthesizeSpeech is the handler for the synthesize_speech tool.
func SynthesizeSpeech(_ context.Context, in contracts.SynthesizeSpeechInput) (tool.Result[contracts.SynthesizeSpeechOutput], error) {
	charCount := len(removePauseMarkers(in.Text))
	estimate := fmt.Sprintf("~%d sec", charCount/20)

	if bifrostClient == nil || SkipLLM {
		return tool.Result[contracts.SynthesizeSpeechOutput]{
			Text: fmt.Sprintf("Synthesized %d characters to %s", charCount, in.OutputPath),
			Structured: contracts.SynthesizeSpeechOutput{
				Kind:             "synthesize",
				OutputPath:       in.OutputPath,
				CharacterCount:   charCount,
				DurationEstimate: estimate,
			},
		}, nil
	}

	format := in.ResponseFormat
	if format == "" {
		format = "mp3"
	}
	// Gemini TTS only supports PCM format
	model := ttsModel()
	if strings.Contains(model, "gemini") && format != "pcm" {
		format = "pcm"
	}

	response, bfErr := bifrostClient.SpeechRequest(
		schemas.NewBifrostContext(context.Background(), schemas.NoDeadline),
		&schemas.BifrostSpeechRequest{
			Provider: schemas.OpenAI,
			Model:    model,
			Input: &schemas.SpeechInput{
				Input: in.Text,
			},
			Params: &schemas.SpeechParameters{
				VoiceConfig: &schemas.SpeechVoiceInput{
					Voice: schemas.Ptr(voiceID(in.Voice)),
				},
				ResponseFormat: format,
			},
		},
	)
	if bfErr != nil {
		return tool.Result[contracts.SynthesizeSpeechOutput]{}, fmt.Errorf("tts failed: %s", bfErr.GetErrorString())
	}

	if response != nil && len(response.Audio) > 0 {
		if err := os.WriteFile(in.OutputPath, response.Audio, 0644); err != nil {
			return tool.Result[contracts.SynthesizeSpeechOutput]{}, fmt.Errorf("failed to write audio: %w", err)
		}
	}

	return tool.Result[contracts.SynthesizeSpeechOutput]{
		Text: fmt.Sprintf("Synthesized %d characters to %s", charCount, in.OutputPath),
		Structured: contracts.SynthesizeSpeechOutput{
			Kind:             "synthesize",
			OutputPath:       in.OutputPath,
			CharacterCount:   charCount,
			DurationEstimate: estimate,
		},
	}, nil
}

// ──────────────────────────────────────────────────────────────────────────────
// Helpers
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
	return "openai/gpt-4o-mini"
}

func ttsModel() string {
	if model := os.Getenv("STUDYAUDIO_TTS_MODEL"); model != "" {
		return model
	}
	if model := os.Getenv("TTS_MODEL"); model != "" {
		return model
	}
	return "tts-1"
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
	return "alloy"
}

func synthesizeToDisk(text, voice, prefix string) (string, error) {
	if bifrostClient == nil {
		return "", fmt.Errorf("tts client not initialized")
	}

	cleanText := removePauseMarkers(text)
	model := ttsModel()

	// Gemini TTS only supports PCM format, convert to mp3 after
	responseFormat := "mp3"
	if strings.Contains(model, "gemini") {
		responseFormat = "pcm"
	}

	response, bfErr := bifrostClient.SpeechRequest(
		schemas.NewBifrostContext(context.Background(), schemas.NoDeadline),
		&schemas.BifrostSpeechRequest{
			Provider: schemas.OpenAI,
			Model:    model,
			Input: &schemas.SpeechInput{
				Input: cleanText,
			},
			Params: &schemas.SpeechParameters{
				VoiceConfig: &schemas.SpeechVoiceInput{
					Voice: schemas.Ptr(voiceID(voice)),
				},
				ResponseFormat: responseFormat,
			},
		},
	)
	if bfErr != nil {
		return "", fmt.Errorf("tts failed: %s", bfErr.GetErrorString())
	}

	if response == nil || len(response.Audio) == 0 {
		return "", fmt.Errorf("no audio returned")
	}

	// For PCM, we need to wrap in a WAV header or save as raw
	// For now, save with appropriate extension
	ext := "mp3"
	if responseFormat == "pcm" {
		ext = "raw" // Raw PCM - caller can convert
	}

	tmpFile, err := os.CreateTemp("", fmt.Sprintf("%s-*.%s", prefix, ext))
	if err != nil {
		return "", err
	}
	defer tmpFile.Close()

	if _, err := tmpFile.Write(response.Audio); err != nil {
		return "", err
	}

	return tmpFile.Name(), nil
}
