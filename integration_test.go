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

func TestIntegrationGenerateStudyGuide(t *testing.T) {
	if os.Getenv("OPENROUTER_API_KEY") == "" {
		t.Skip("OPENROUTER_API_KEY not set, skipping integration test")
	}

	in := contracts.GenerateStudyGuideInput{
		Content:        "The cutaneous immune system: skin as an active immunological organ. Innate vs adaptive immunity. Keratinocyte as central sentinel cell. TLR receptors and inflammasome. UV radiation effects.",
		Language:       "es",
		Difficulty:     "masters",
		DurationTarget: "short",
		PreviewOnly:    true,
	}

	res, err := handlers.GenerateStudyGuide(context.Background(), in)
	if err != nil {
		t.Fatalf("GenerateStudyGuide failed: %v", err)
	}

	fmt.Printf("=== STUDY GUIDE ===\n")
	fmt.Printf("Word count: %d\n", res.Structured.WordCount)
	fmt.Printf("Duration: %s\n", res.Structured.DurationEstimate)
	fmt.Printf("Sections: %d\n", len(res.Structured.Sections))

	// Check for audio tags
	script := res.Structured.Script
	tags := []string{"[warm]", "[thoughtful]", "[normal voice]", "[curious]", "[emphasizing]", "[serious]"}
	foundTags := 0
	for _, tag := range tags {
		if contains(script, tag) {
			foundTags++
			fmt.Printf("Found tag: %s\n", tag)
		}
	}
	fmt.Printf("Audio tags found: %d/%d\n", foundTags, len(tags))

	// Print first 500 chars
	if len(script) > 500 {
		fmt.Printf("Script preview:\n%s...\n", script[:500])
	} else {
		fmt.Printf("Script:\n%s\n", script)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
