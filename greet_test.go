package main

import (
	"context"
	"testing"

	"github.com/hurtener/go-study-mcp/internal/contracts"
	"github.com/hurtener/go-study-mcp/internal/handlers"
)

func TestGeneratePodcast(t *testing.T) {
	tests := []struct {
		name     string
		in       contracts.GeneratePodcastInput
		wantLang string
	}{
		{
			name: "english default",
			in: contracts.GeneratePodcastInput{
				Content: "The mitochondria is the powerhouse of the cell.",
			},
			wantLang: "en",
		},
		{
			name: "spanish",
			in: contracts.GeneratePodcastInput{
				Content: "La mitocondria es la powerhouse de la célula.",
				Language: "es",
			},
			wantLang: "es",
		},
		{
			name: "preview only",
			in: contracts.GeneratePodcastInput{
				Content:    "Test content",
				PreviewOnly: true,
			},
			wantLang: "en",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := handlers.GeneratePodcast(context.Background(), tt.in)
			if err != nil {
				t.Fatalf("GeneratePodcast: %v", err)
			}
			if res.Structured.Language != tt.wantLang {
				t.Errorf("Language = %q, want %q", res.Structured.Language, tt.wantLang)
			}
			if res.Structured.Script == "" {
				t.Error("Script should not be empty")
			}
			if res.Structured.WordCount == 0 {
				t.Error("WordCount should be > 0")
			}
		})
	}
}

func TestGenerateFlashcards(t *testing.T) {
	t.Run("preview with cards", func(t *testing.T) {
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
		if res.Structured.CardCount != 2 {
			t.Errorf("CardCount = %d, want 2", res.Structured.CardCount)
		}
		if !res.Structured.PreviewOnly {
			t.Error("PreviewOnly should be true")
		}
	})

	t.Run("no content or cards", func(t *testing.T) {
		in := contracts.GenerateFlashcardsInput{}
		_, err := handlers.GenerateFlashcards(context.Background(), in)
		if err == nil {
			t.Error("expected error when neither content nor cards provided")
		}
	})
}

func TestSynthesizeSpeech(t *testing.T) {
	in := contracts.SynthesizeSpeechInput{
		Text:       "Hello world [PAUSE:2] this is a test.",
		OutputPath: "/tmp/test.mp3",
	}
	res, err := handlers.SynthesizeSpeech(context.Background(), in)
	if err != nil {
		t.Fatalf("SynthesizeSpeech: %v", err)
	}
	if res.Structured.OutputPath != "/tmp/test.mp3" {
		t.Errorf("OutputPath = %q, want %q", res.Structured.OutputPath, "/tmp/test.mp3")
	}
	if res.Structured.CharacterCount == 0 {
		t.Error("CharacterCount should be > 0")
	}
}
