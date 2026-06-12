// Package handlers implements the tool handler functions for go-study-mcp.
//
// Each handler receives decoded, schema-valid input and returns a typed
// tool.Result that splits model-facing text from UI-facing structured data.
package handlers

import (
	"context"
	"fmt"
	"strings"

	"github.com/hurtener/dockyard/runtime/tool"

	"github.com/hurtener/go-study-mcp/internal/contracts"
	"github.com/hurtener/go-study-mcp/internal/prompts"
)

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

	// TODO: Replace with actual LLM call using system + user prompts.
	// For now, return the prompts as the script for validation.
	script := fmt.Sprintf("[SYSTEM]\n%s\n\n[USER]\n%s", system, user)

	if in.PreviewOnly {
		return tool.Result[contracts.GeneratePodcastOutput]{
			Text: script,
			Structured: contracts.GeneratePodcastOutput{
				Script:           script,
				WordCount:        len(strings.Fields(script)),
				DurationEstimate: estimateDuration(wordTarget),
				Language:         lang,
				PreviewOnly:      true,
			},
		}, nil
	}

	// TODO: Call TTS engine with the generated script.
	// For now, return preview-only output.
	return tool.Result[contracts.GeneratePodcastOutput]{
		Text: script,
		Structured: contracts.GeneratePodcastOutput{
			Script:           script,
			WordCount:        len(strings.Fields(script)),
			DurationEstimate: estimateDuration(wordTarget),
			Language:         lang,
			PreviewOnly:      true,
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
		// Pre-curated cards: skip LLM entirely.
		cards = in.Cards
	} else if in.Content != "" {
		// Content mode: generate cards via LLM.
		// TODO: Replace with actual LLM call.
		system, user := prompts.FlashcardExtract(prompts.FlashcardParams{
			Language:   lang,
			Difficulty: in.Difficulty,
			CardCount:  in.CardCount,
			Content:    in.Content,
		})
		_ = system
		_ = user

		// Placeholder: return empty cards for validation.
		cards = []contracts.FlashCard{}
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
				Cards:            cards,
				CardCount:        cardCount,
				DurationEstimate: estimate,
				Language:         lang,
				PreviewOnly:      true,
			},
		}, nil
	}

	// TODO: Call TTS engine with assembled flashcard script.
	return tool.Result[contracts.GenerateFlashcardsOutput]{
		Text: formatCardsText(cards),
		Structured: contracts.GenerateFlashcardsOutput{
			Cards:            cards,
			CardCount:        cardCount,
			DurationEstimate: estimate,
			Language:         lang,
			PreviewOnly:      true,
		},
	}, nil
}

// SynthesizeSpeech is the handler for the synthesize_speech tool.
func SynthesizeSpeech(_ context.Context, in contracts.SynthesizeSpeechInput) (tool.Result[contracts.SynthesizeSpeechOutput], error) {
	text := in.Text
	format := in.ResponseFormat
	if format == "" {
		format = "mp3"
	}

	// Count actual speech characters (exclude [PAUSE:N] markers).
	charCount := len(removePauseMarkers(text))

	// TODO: Replace with actual TTS call.
	estimate := fmt.Sprintf("~%d sec", charCount/20)

	return tool.Result[contracts.SynthesizeSpeechOutput]{
		Text: fmt.Sprintf("Synthesized %d characters to %s", charCount, in.OutputPath),
		Structured: contracts.SynthesizeSpeechOutput{
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
	default: // "medium" or unset
		return 700
	}
}

func estimateDuration(words int) string {
	minutes := words / 150 // ~150 wpm
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

func removePauseMarkers(text string) string {
	// Simple implementation: remove [PAUSE:N] patterns.
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
