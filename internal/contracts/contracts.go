// Package contracts holds this server's tool input and output contracts.
//
// These typed Go structs are the SOURCE OF TRUTH for each tool's schema
// (Dockyard P1 — contract-first, RFC §6). The JSON Schema and TypeScript
// alongside this file are GENERATED from these structs by `dockyard generate`;
// never hand-edit a generated file. Change a contract here, then regenerate.
package contracts

// ──────────────────────────────────────────────────────────────────────────────
// generate_podcast
// ──────────────────────────────────────────────────────────────────────────────

// GeneratePodcastInput is the typed input for the generate_podcast tool.
type GeneratePodcastInput struct {
	// Content is the raw study material to transform into a narrated podcast. Required.
	Content string `json:"content"`
	// Language is the ISO 639-1 code for the output (e.g. "en", "es", "fr"). Required.
	Language string `json:"language"`
	// DurationTarget controls approximate output length: "short" (~2 min),
	// "medium" (~5 min), or "long" (~10 min). Optional; defaults to "medium".
	DurationTarget string `json:"durationTarget,omitempty"`
	// Tone sets the narration style: "casual", "academic", "enthusiastic", or "calm".
	// Optional; defaults to "casual".
	Tone string `json:"tone,omitempty"`
	// Persona is a free-form string that overrides the narrator identity
	// (e.g. "a friendly tutor", "a professor specializing in biology").
	// When set, it takes precedence over Tone. Optional.
	Persona string `json:"persona,omitempty"`
	// TopicHint is a short label for the script framing (e.g. "Krebs Cycle").
	// Optional; helps the LLM focus the narration.
	TopicHint string `json:"topicHint,omitempty"`
	// Voice is the TTS voice identifier. Optional; provider-specific default
	// is used when empty.
	Voice string `json:"voice,omitempty"`
	// PreviewOnly when true returns the generated script without calling TTS.
	// Useful for reviewing before committing to audio. Optional; defaults to false.
	PreviewOnly bool `json:"previewOnly,omitempty"`
}

// GeneratePodcastOutput is the typed output for the generate_podcast tool.
type GeneratePodcastOutput struct {
	// Kind discriminates the output type for the UI dispatcher.
	Kind string `json:"kind"`
	// Script is the full narration text produced by the LLM.
	Script string `json:"script"`
	// OutputPath is the absolute path to the generated audio file.
	// Empty when PreviewOnly is true or while the job is still processing.
	OutputPath string `json:"outputPath,omitempty"`
	// JobID identifies the background generation job. Set when PreviewOnly is
	// false; poll list_jobs / read_audio with it. Empty in preview mode.
	JobID string `json:"jobId,omitempty"`
	// Status is the job lifecycle state at return time (e.g. "processing").
	// Empty in preview mode.
	Status string `json:"status,omitempty"`
	// WordCount is the number of words in the generated script.
	WordCount int `json:"wordCount"`
	// DurationEstimate is a human-readable approximation (e.g. "~5 min").
	DurationEstimate string `json:"durationEstimate"`
	// Language echoes the requested language code.
	Language string `json:"language"`
	// PreviewOnly echoes the preview flag.
	PreviewOnly bool `json:"previewOnly"`
}

// ──────────────────────────────────────────────────────────────────────────────
// generate_flashcards
// ──────────────────────────────────────────────────────────────────────────────

// FlashCard is a single question-answer pair.
type FlashCard struct {
	// Question is the recall prompt.
	Question string `json:"question"`
	// Answer is the expected response.
	Answer string `json:"answer"`
}

// GenerateFlashcardsInput is the typed input for the generate_flashcards tool.
type GenerateFlashcardsInput struct {
	// Content is the raw study material for LLM extraction.
	// Mutually exclusive with Cards. At least one must be provided.
	Content string `json:"content,omitempty"`
	// Cards is a pre-curated list of Q&A pairs. When provided the LLM is
	// skipped entirely. Mutually exclusive with Content.
	Cards []FlashCard `json:"cards,omitempty"`
	// Language is the ISO 639-1 code for the cards (e.g. "en", "es").
	// Ignored when Cards is provided. Optional; defaults to "en".
	Language string `json:"language,omitempty"`
	// Difficulty calibrates question depth: "basic" (definition/recall),
	// "intermediate" (comprehension/comparison), or "advanced"
	// (application/synthesis). Ignored when Cards is provided.
	// Optional; defaults to "intermediate".
	Difficulty string `json:"difficulty,omitempty"`
	// CardCount is the number of Q&A pairs to generate (3–40).
	// Ignored when Cards is provided. Optional; defaults to 10.
	CardCount int `json:"cardCount,omitempty"`
	// PauseDuration is the seconds of silence between question and answer
	// in the audio output (2–30). Optional; defaults to 5.
	PauseDuration int `json:"pauseDuration,omitempty"`
	// Voice is the TTS voice identifier. Optional.
	Voice string `json:"voice,omitempty"`
	// PreviewOnly when true returns the card list without audio synthesis.
	// Optional; defaults to false.
	PreviewOnly bool `json:"previewOnly,omitempty"`
}

// GenerateFlashcardsOutput is the typed output for the generate_flashcards tool.
type GenerateFlashcardsOutput struct {
	// Kind discriminates the output type for the UI dispatcher.
	Kind string `json:"kind"`
	// Cards is the list of generated or provided Q&A pairs.
	Cards []FlashCard `json:"cards"`
	// OutputPath is the absolute path to the generated audio file.
	// Empty when PreviewOnly is true or while the job is still processing.
	OutputPath string `json:"outputPath,omitempty"`
	// JobID identifies the background generation job. Set when PreviewOnly is
	// false; poll list_jobs / read_audio with it. Empty in preview mode.
	JobID string `json:"jobId,omitempty"`
	// Status is the job lifecycle state at return time (e.g. "processing").
	// Empty in preview mode.
	Status string `json:"status,omitempty"`
	// CardCount is the actual number of cards in the output.
	CardCount int `json:"cardCount"`
	// DurationEstimate is a human-readable approximation (e.g. "~6 min").
	DurationEstimate string `json:"durationEstimate"`
	// Language echoes the language code used for generation.
	Language string `json:"language"`
	// PreviewOnly echoes the preview flag.
	PreviewOnly bool `json:"previewOnly"`
}

// ──────────────────────────────────────────────────────────────────────────────
// synthesize_speech
// ──────────────────────────────────────────────────────────────────────────────

// SynthesizeSpeechInput is the typed input for the synthesize_speech tool.
//
// Note: the destination path is owned by the server (a writable OUTPUT_DIR
// resolved at startup), never supplied by the caller — a caller-chosen path
// is unreliable across hosts (read-only CWD, non-existent mounts).
type SynthesizeSpeechInput struct {
	// Text is the exact text to synthesize. May include [PAUSE:N] markers
	// for timed silence (N = seconds). Required.
	Text string `json:"text"`
	// Voice is the TTS voice identifier. Optional.
	Voice string `json:"voice,omitempty"`
	// ResponseFormat selects the audio container: "mp3" or "wav".
	// Optional; defaults to "mp3".
	ResponseFormat string `json:"responseFormat,omitempty"`
}

// SynthesizeSpeechOutput is the typed output for the synthesize_speech tool.
type SynthesizeSpeechOutput struct {
	// Kind discriminates the output type for the UI dispatcher.
	Kind string `json:"kind"`
	// OutputPath is the absolute path to the generated audio file.
	// Empty while the job is still processing.
	OutputPath string `json:"outputPath,omitempty"`
	// JobID identifies the background synthesis job; poll list_jobs /
	// read_audio with it.
	JobID string `json:"jobId,omitempty"`
	// Status is the job lifecycle state at return time (e.g. "processing").
	Status string `json:"status,omitempty"`
	// CharacterCount is the number of characters sent to the TTS engine
	// (excludes [PAUSE:N] marker text).
	CharacterCount int `json:"characterCount"`
	// DurationEstimate is a human-readable approximation (e.g. "~3 min").
	DurationEstimate string `json:"durationEstimate"`
}

// ──────────────────────────────────────────────────────────────────────────────
// generate_study_guide
// ──────────────────────────────────────────────────────────────────────────────

// StudyGuideSection is a single section of the generated study guide.
type StudyGuideSection struct {
	// Title is the section heading.
	Title string `json:"title"`
	// Content is the detailed explanation text.
	Content string `json:"content"`
}

// GenerateStudyGuideInput is the typed input for the generate_study_guide tool.
type GenerateStudyGuideInput struct {
	// Content is the raw study material to transform into a deep study guide. Required.
	Content string `json:"content"`
	// Language is the ISO 639-1 code for the output (e.g. "en", "es"). Required.
	Language string `json:"language"`
	// Difficulty is the target academic level: "undergraduate", "graduate",
	// "masters", or "phd". Controls depth and detail level. Optional; defaults to "graduate".
	Difficulty string `json:"difficulty,omitempty"`
	// DurationTarget controls approximate audio output length: "short" (~5 min),
	// "medium" (~15 min), or "long" (~30 min). Optional; defaults to "medium".
	DurationTarget string `json:"durationTarget,omitempty"`
	// Voice is the TTS voice identifier. Optional.
	Voice string `json:"voice,omitempty"`
	// PreviewOnly when true returns the study guide text without audio synthesis.
	// Optional; defaults to false.
	PreviewOnly bool `json:"previewOnly,omitempty"`
}

// GenerateStudyGuideOutput is the typed output for the generate_study_guide tool.
type GenerateStudyGuideOutput struct {
	// Kind discriminates the output type for the UI dispatcher.
	Kind string `json:"kind"`
	// Script is the full narrated study guide text.
	Script string `json:"script"`
	// Sections are the structured sections extracted from the script.
	Sections []StudyGuideSection `json:"sections"`
	// OutputPath is the absolute path to the generated audio file.
	// Empty when PreviewOnly is true or while the job is still processing.
	OutputPath string `json:"outputPath,omitempty"`
	// JobID identifies the background generation job. Set when PreviewOnly is
	// false; poll list_jobs / read_audio with it. Empty in preview mode.
	JobID string `json:"jobId,omitempty"`
	// Status is the job lifecycle state at return time (e.g. "processing").
	// Empty in preview mode.
	Status string `json:"status,omitempty"`
	// WordCount is the total word count.
	WordCount int `json:"wordCount"`
	// DurationEstimate is a human-readable approximation (e.g. "~15 min").
	DurationEstimate string `json:"durationEstimate"`
	// Language echoes the requested language code.
	Language string `json:"language"`
	// Difficulty echoes the requested difficulty level.
	Difficulty string `json:"difficulty"`
	// PreviewOnly echoes the preview flag.
	PreviewOnly bool `json:"previewOnly"`
}

// ──────────────────────────────────────────────────────────────────────────────
// list_jobs
// ──────────────────────────────────────────────────────────────────────────────

// Job is a single asynchronous audio-generation job, as surfaced to the UI.
type Job struct {
	// ID is the unique job identifier (pass to read_audio).
	ID string `json:"id"`
	// Kind is the originating tool: "podcast", "study_guide", "flashcards",
	// or "synthesize".
	Kind string `json:"kind"`
	// Title is a short human-readable label (topic hint or text snippet).
	Title string `json:"title"`
	// Status is the lifecycle state: "queued", "processing", "completed",
	// or "failed".
	Status string `json:"status"`
	// CreatedAt is the RFC3339 creation timestamp.
	CreatedAt string `json:"createdAt"`
	// UpdatedAt is the RFC3339 timestamp of the last state change.
	UpdatedAt string `json:"updatedAt"`
	// OutputPath is the absolute path to the generated audio file. Set only
	// when Status is "completed".
	OutputPath string `json:"outputPath,omitempty"`
	// DurationEstimate is a human-readable approximation (e.g. "~5 min").
	DurationEstimate string `json:"durationEstimate,omitempty"`
	// CharacterCount is the number of characters synthesized, when known.
	CharacterCount int `json:"characterCount,omitempty"`
	// WordCount is the script word count, when known.
	WordCount int `json:"wordCount,omitempty"`
	// Error is the failure reason. Set only when Status is "failed".
	Error string `json:"error,omitempty"`
}

// ListJobsInput is the (empty) input for the list_jobs tool.
type ListJobsInput struct{}

// ListJobsOutput is the typed output for the list_jobs tool.
type ListJobsOutput struct {
	// Kind discriminates the output type for the UI dispatcher.
	Kind string `json:"kind"`
	// Jobs is the list of all jobs, newest first.
	Jobs []Job `json:"jobs"`
}

// ──────────────────────────────────────────────────────────────────────────────
// list_voices
// ──────────────────────────────────────────────────────────────────────────────

// Voice is a single TTS voice option for the active provider.
type Voice struct {
	// ID is the voice identifier passed to the TTS engine.
	ID string `json:"id"`
	// Label is the human-readable display name.
	Label string `json:"label"`
	// Description is a short characteristic (e.g. "Warm", "Clear").
	Description string `json:"description,omitempty"`
}

// ListVoicesInput is the (empty) input for the list_voices tool.
type ListVoicesInput struct{}

// ListVoicesOutput is the typed output for the list_voices tool.
type ListVoicesOutput struct {
	// Kind discriminates the output type for the UI dispatcher.
	Kind string `json:"kind"`
	// Provider is the active TTS provider ("gemini" or "openai").
	Provider string `json:"provider"`
	// Model is the configured TTS model id.
	Model string `json:"model"`
	// DefaultVoice is the voice used when none is specified.
	DefaultVoice string `json:"defaultVoice"`
	// Voices are the voices the active provider can speak.
	Voices []Voice `json:"voices"`
}

// ──────────────────────────────────────────────────────────────────────────────
// read_audio
// ──────────────────────────────────────────────────────────────────────────────

// ReadAudioInput is the typed input for the read_audio tool. Provide either
// JobID (preferred) or Path; Path is confined to the server's OUTPUT_DIR.
type ReadAudioInput struct {
	// JobID is the job whose completed audio to read.
	JobID string `json:"jobId,omitempty"`
	// Path is an explicit audio file path under OUTPUT_DIR.
	Path string `json:"path,omitempty"`
}

// ReadAudioOutput is the typed output for the read_audio tool.
type ReadAudioOutput struct {
	// Kind discriminates the output type for the UI dispatcher.
	Kind string `json:"kind"`
	// DataURI is "data:<mime>;base64,<...>" — the UI converts it to a blob:
	// URL for inline <audio> playback. Empty when Truncated is true.
	DataURI string `json:"dataUri,omitempty"`
	// Mime is the audio MIME type (e.g. "audio/mpeg").
	Mime string `json:"mime,omitempty"`
	// SizeBytes is the file size on disk.
	SizeBytes int64 `json:"sizeBytes,omitempty"`
	// Truncated is true when the file exceeds the inline cap; the UI falls
	// back to showing the OutputPath rather than shipping a huge payload.
	Truncated bool `json:"truncated,omitempty"`
}
