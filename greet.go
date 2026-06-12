package main

import (
	"github.com/hurtener/dockyard/runtime/server"
	"github.com/hurtener/dockyard/runtime/tool"

	"github.com/hurtener/go-study-mcp/internal/contracts"
	"github.com/hurtener/go-study-mcp/internal/handlers"
)

// registerTools declares and registers every tool this server exposes.
func registerTools(srv *server.Server) error {
	if err := tool.New[contracts.GeneratePodcastInput, contracts.GeneratePodcastOutput]("generate_podcast").
		Describe("Transform study material into a narrated podcast-style audio with flexible tone, persona, and language support.").
		UI(appName).
		Handler(handlers.GeneratePodcast).
		Register(srv); err != nil {
		return err
	}

	if err := tool.New[contracts.GenerateFlashcardsInput, contracts.GenerateFlashcardsOutput]("generate_flashcards").
		Describe("Generate Q&A flashcard audio from study material with timed pauses for active recall.").
		UI(appName).
		Handler(handlers.GenerateFlashcards).
		Register(srv); err != nil {
		return err
	}

	if err := tool.New[contracts.SynthesizeSpeechInput, contracts.SynthesizeSpeechOutput]("synthesize_speech").
		Describe("Direct text-to-speech synthesis with support for [PAUSE:N] markers.").
		UI(appName).
		Handler(handlers.SynthesizeSpeech).
		Register(srv); err != nil {
		return err
	}

	return nil
}
