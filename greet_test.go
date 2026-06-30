package main

import (
	"context"
	"os"
	"testing"

	"github.com/hurtener/go-study-mcp/internal/contracts"
	"github.com/hurtener/go-study-mcp/internal/handlers"
)

func TestMain(m *testing.M) {
	// Set a placeholder API key for tests (handlers will fail LLM calls but structure is valid)
	os.Setenv("OPENAI_API_KEY", "sk-test")
	os.Exit(m.Run())
}

func TestGeneratePodcast(t *testing.T) {
	t.Run("preview only returns script", func(t *testing.T) {
		in := contracts.GeneratePodcastInput{
			Content:     "The mitochondria is the powerhouse of the cell.",
			Language:    "en",
			PreviewOnly: true,
		}
		res, err := handlers.GeneratePodcast(context.Background(), in)
		if err != nil {
			t.Fatalf("GeneratePodcast: %v", err)
		}
		if res.Structured.Kind != "podcast" {
			t.Errorf("Kind = %q, want %q", res.Structured.Kind, "podcast")
		}
		if res.Structured.Language != "en" {
			t.Errorf("Language = %q, want %q", res.Structured.Language, "en")
		}
		if res.Structured.Script == "" {
			t.Error("Script should not be empty")
		}
		if !res.Structured.PreviewOnly {
			t.Error("PreviewOnly should be true")
		}
	})

	t.Run("spanish language", func(t *testing.T) {
		in := contracts.GeneratePodcastInput{
			Content:     "La mitocondria es la powerhouse de la célula.",
			Language:    "es",
			PreviewOnly: true,
		}
		res, err := handlers.GeneratePodcast(context.Background(), in)
		if err != nil {
			t.Fatalf("GeneratePodcast: %v", err)
		}
		if res.Structured.Language != "es" {
			t.Errorf("Language = %q, want %q", res.Structured.Language, "es")
		}
	})
}

func TestGenerateFlashcards(t *testing.T) {
	t.Run("preview with pre-curated cards", func(t *testing.T) {
		in := contracts.GenerateFlashcardsInput{
			Cards: []contracts.FlashCard{
				{Question: "What is 2+2?", Answer: "4"},
				{Question: "What is the capital of France?", Answer: "Paris"},
			},
			PreviewOnly: true,
		}
		res, err := handlers.GenerateFlashcards(context.Background(), in)
		if err != nil {
			t.Fatalf("GenerateFlashcards: %v", err)
		}
		if res.Structured.Kind != "flashcards" {
			t.Errorf("Kind = %q, want %q", res.Structured.Kind, "flashcards")
		}
		if res.Structured.CardCount != 2 {
			t.Errorf("CardCount = %d, want 2", res.Structured.CardCount)
		}
		if !res.Structured.PreviewOnly {
			t.Error("PreviewOnly should be true")
		}
		if len(res.Structured.Cards) != 2 {
			t.Errorf("Cards len = %d, want 2", len(res.Structured.Cards))
		}
	})

	t.Run("no content or cards returns error", func(t *testing.T) {
		in := contracts.GenerateFlashcardsInput{}
		_, err := handlers.GenerateFlashcards(context.Background(), in)
		if err == nil {
			t.Error("expected error when neither content nor cards provided")
		}
	})
}

func TestSynthesizeSpeech(t *testing.T) {
	// With a placeholder key (TestMain sets sk-test) TTS is not configured, so
	// the handler returns a failed status synchronously rather than enqueuing a
	// job. The character count excludes [PAUSE:N] marker text.
	in := contracts.SynthesizeSpeechInput{
		Text: "Hello world [PAUSE:2] this is a test.",
	}
	res, err := handlers.SynthesizeSpeech(context.Background(), in)
	if err != nil {
		t.Fatalf("SynthesizeSpeech: %v", err)
	}
	if res.Structured.Kind != "synthesize" {
		t.Errorf("Kind = %q, want %q", res.Structured.Kind, "synthesize")
	}
	if res.Structured.CharacterCount == 0 {
		t.Error("CharacterCount should be > 0")
	}
}
