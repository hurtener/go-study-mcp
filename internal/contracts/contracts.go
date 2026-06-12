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
	// Empty when PreviewOnly is true.
	OutputPath string `json:"outputPath,omitempty"`
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
	// Empty when PreviewOnly is true.
	OutputPath string `json:"outputPath,omitempty"`
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
type SynthesizeSpeechInput struct {
	// Text is the exact text to synthesize. May include [PAUSE:N] markers
	// for timed silence (N = seconds). Required.
	Text string `json:"text"`
	// OutputPath is the destination file path. Required.
	OutputPath string `json:"outputPath"`
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
	OutputPath string `json:"outputPath"`
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
	// Empty when PreviewOnly is true.
	OutputPath string `json:"outputPath,omitempty"`
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
