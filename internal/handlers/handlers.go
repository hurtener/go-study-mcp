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

	openai "github.com/sashabaranov/go-openai"

	"github.com/hurtener/dockyard/runtime/tool"

	"github.com/hurtener/go-study-mcp/internal/contracts"
	"github.com/hurtener/go-study-mcp/internal/prompts"
)

// llmClient is the shared OpenAI-compatible client.
// Configure via OPENAI_API_KEY and OPENAI_BASE_URL (for Bifrost/gateways).
var llmClient *openai.Client

// SkipLLM is a test flag that skips actual LLM calls.
var SkipLLM bool

func init() {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		apiKey = "sk-placeholder"
		SkipLLM = true
	}

	cfg := openai.DefaultConfig(apiKey)
	if baseURL := os.Getenv("OPENAI_BASE_URL"); baseURL != "" {
		cfg.BaseURL = baseURL
	}

	llmClient = openai.NewClientWithConfig(cfg)
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
	if SkipLLM {
		// In test mode, return the prompts as the script
		script = fmt.Sprintf("[SYSTEM]\n%s\n\n[USER]\n%s", system, user)
	} else {
		resp, err := llmClient.CreateChatCompletion(context.Background(), openai.ChatCompletionRequest{
			Model: modelName(),
			Messages: []openai.ChatCompletionMessage{
				{Role: openai.ChatMessageRoleSystem, Content: system},
				{Role: openai.ChatMessageRoleUser, Content: user},
			},
			Temperature: 0.7,
			MaxTokens:   wordTarget * 2,
		})
		if err != nil {
			return tool.Result[contracts.GeneratePodcastOutput]{}, fmt.Errorf("llm call failed: %w", err)
		}
		script = resp.Choices[0].Message.Content
	}
	wordCount := len(strings.Fields(script))

	if in.PreviewOnly {
		return tool.Result[contracts.GeneratePodcastOutput]{
			Text: script,
			Structured: contracts.GeneratePodcastOutput{
				Kind:              "podcast",
				Script:            script,
				WordCount:         wordCount,
				DurationEstimate:  estimateDuration(wordCount),
				Language:          lang,
				PreviewOnly:       true,
			},
		}, nil
	}

	// TODO: TTS synthesis via Bifrost SpeechRequest
	// For now, return preview-only output
	return tool.Result[contracts.GeneratePodcastOutput]{
		Text: script,
		Structured: contracts.GeneratePodcastOutput{
			Kind:              "podcast",
			Script:            script,
			WordCount:         wordCount,
			DurationEstimate:  estimateDuration(wordCount),
			Language:          lang,
			PreviewOnly:       true,
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

		if SkipLLM {
			// In test mode, return empty cards
			cards = []contracts.FlashCard{}
		} else {
			resp, err := llmClient.CreateChatCompletion(context.Background(), openai.ChatCompletionRequest{
				Model: modelName(),
				Messages: []openai.ChatCompletionMessage{
					{Role: openai.ChatMessageRoleSystem, Content: system},
					{Role: openai.ChatMessageRoleUser, Content: user},
				},
				Temperature: 0.3,
				MaxTokens:   4096,
			})
			if err != nil {
				return tool.Result[contracts.GenerateFlashcardsOutput]{}, fmt.Errorf("llm call failed: %w", err)
			}

			raw := resp.Choices[0].Message.Content
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
				Kind:              "flashcards",
				Cards:             cards,
				CardCount:         cardCount,
				DurationEstimate:  estimate,
				Language:          lang,
				PreviewOnly:       true,
			},
		}, nil
	}

	// TODO: TTS synthesis
	return tool.Result[contracts.GenerateFlashcardsOutput]{
		Text: formatCardsText(cards),
		Structured: contracts.GenerateFlashcardsOutput{
			Kind:              "flashcards",
			Cards:             cards,
			CardCount:         cardCount,
			DurationEstimate:  estimate,
			Language:          lang,
			PreviewOnly:       true,
		},
	}, nil
}

// SynthesizeSpeech is the handler for the synthesize_speech tool.
func SynthesizeSpeech(_ context.Context, in contracts.SynthesizeSpeechInput) (tool.Result[contracts.SynthesizeSpeechOutput], error) {
	charCount := len(removePauseMarkers(in.Text))
	estimate := fmt.Sprintf("~%d sec", charCount/20)

	return tool.Result[contracts.SynthesizeSpeechOutput]{
		Text: fmt.Sprintf("Synthesized %d characters to %s", charCount, in.OutputPath),
		Structured: contracts.SynthesizeSpeechOutput{
			Kind:            "synthesize",
			OutputPath:      in.OutputPath,
			CharacterCount:  charCount,
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

func modelName() string {
	if model := os.Getenv("OPENAI_MODEL"); model != "" {
		return model
	}
	return openai.GPT4oMini
}
