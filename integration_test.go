//go:build integration

package main

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/hurtener/go-study-mcp/internal/contracts"
	"github.com/hurtener/go-study-mcp/internal/handlers"
)

func TestIntegrationGeneratePodcast(t *testing.T) {
	if os.Getenv("OPENROUTER_API_KEY") == "" {
		t.Skip("OPENROUTER_API_KEY not set, skipping integration test")
	}

	in := contracts.GeneratePodcastInput{
		Content:        "The mitochondria is the powerhouse of the cell. It produces ATP through cellular respiration.",
		Language:       "en",
		DurationTarget: "short",
		Tone:           "casual",
		PreviewOnly:    true,
	}

	res, err := handlers.GeneratePodcast(context.Background(), in)
	if err != nil {
		t.Fatalf("GeneratePodcast failed: %v", err)
	}

	fmt.Printf("=== PODCAST SCRIPT ===\n%s\n", res.Structured.Script)
	fmt.Printf("Word count: %d\n", res.Structured.WordCount)
	fmt.Printf("Duration: %s\n", res.Structured.DurationEstimate)
}

func TestIntegrationGenerateFlashcards(t *testing.T) {
	if os.Getenv("OPENROUTER_API_KEY") == "" {
		t.Skip("OPENROUTER_API_KEY not set, skipping integration test")
	}

	in := contracts.GenerateFlashcardsInput{
		Content:     "The mitochondria is the powerhouse of the cell. It produces ATP through cellular respiration.",
		Language:    "en",
		Difficulty:  "intermediate",
		CardCount:   3,
		PreviewOnly: true,
	}

	res, err := handlers.GenerateFlashcards(context.Background(), in)
	if err != nil {
		t.Fatalf("GenerateFlashcards failed: %v", err)
	}

	fmt.Printf("=== FLASHCARDS ===\n")
	for i, c := range res.Structured.Cards {
		fmt.Printf("[%d] Q: %s\n    A: %s\n\n", i+1, c.Question, c.Answer)
	}
	fmt.Printf("Card count: %d\n", res.Structured.CardCount)
}

func TestIntegrationSynthesizeSpeech(t *testing.T) {
	if os.Getenv("OPENROUTER_API_KEY") == "" {
		t.Skip("OPENROUTER_API_KEY not set, skipping integration test")
	}

	in := contracts.SynthesizeSpeechInput{
		Text:       "Hello! This is a test of the study audio synthesis system.",
		OutputPath: "/tmp/test_synth.mp3",
	}

	res, err := handlers.SynthesizeSpeech(context.Background(), in)
	if err != nil {
		t.Fatalf("SynthesizeSpeech failed: %v", err)
	}

	fmt.Printf("=== SYNTHESIZE ===\n")
	fmt.Printf("Output: %s\n", res.Structured.OutputPath)
	fmt.Printf("Characters: %d\n", res.Structured.CharacterCount)

	if res.Structured.OutputPath != "" {
		if info, err := os.Stat(res.Structured.OutputPath); err == nil {
			fmt.Printf("File size: %d bytes\n", info.Size())
		}
	}
}
