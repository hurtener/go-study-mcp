package handlers

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/hurtener/dockyard/runtime/tool"

	"github.com/hurtener/go-study-mcp/internal/contracts"
)

// Voice catalogs per TTS provider. The active provider is derived from the
// configured TTS model, so the UI (and the model) only ever see voices that
// the model can actually speak.

// geminiVoices are the prebuilt voices for Google Gemini flash TTS, each with
// its documented characteristic.
var geminiVoices = []contracts.Voice{
	{ID: "Erinome", Label: "Erinome", Description: "Clear"},
	{ID: "Zephyr", Label: "Zephyr", Description: "Bright"},
	{ID: "Puck", Label: "Puck", Description: "Upbeat"},
	{ID: "Charon", Label: "Charon", Description: "Informative"},
	{ID: "Kore", Label: "Kore", Description: "Firm"},
	{ID: "Fenrir", Label: "Fenrir", Description: "Excitable"},
	{ID: "Leda", Label: "Leda", Description: "Youthful"},
	{ID: "Orus", Label: "Orus", Description: "Firm"},
	{ID: "Aoede", Label: "Aoede", Description: "Breezy"},
	{ID: "Callirrhoe", Label: "Callirrhoe", Description: "Easy-going"},
	{ID: "Autonoe", Label: "Autonoe", Description: "Bright"},
	{ID: "Enceladus", Label: "Enceladus", Description: "Breathy"},
	{ID: "Iapetus", Label: "Iapetus", Description: "Clear"},
	{ID: "Umbriel", Label: "Umbriel", Description: "Easy-going"},
	{ID: "Algieba", Label: "Algieba", Description: "Smooth"},
	{ID: "Despina", Label: "Despina", Description: "Smooth"},
	{ID: "Algenib", Label: "Algenib", Description: "Gravelly"},
	{ID: "Rasalgethi", Label: "Rasalgethi", Description: "Informative"},
	{ID: "Laomedeia", Label: "Laomedeia", Description: "Upbeat"},
	{ID: "Achernar", Label: "Achernar", Description: "Soft"},
	{ID: "Alnilam", Label: "Alnilam", Description: "Firm"},
	{ID: "Schedar", Label: "Schedar", Description: "Even"},
	{ID: "Gacrux", Label: "Gacrux", Description: "Mature"},
	{ID: "Pulcherrima", Label: "Pulcherrima", Description: "Forward"},
	{ID: "Achird", Label: "Achird", Description: "Friendly"},
	{ID: "Zubenelgenubi", Label: "Zubenelgenubi", Description: "Casual"},
	{ID: "Vindemiatrix", Label: "Vindemiatrix", Description: "Gentle"},
	{ID: "Sadachbia", Label: "Sadachbia", Description: "Lively"},
	{ID: "Sadaltager", Label: "Sadaltager", Description: "Knowledgeable"},
	{ID: "Sulafat", Label: "Sulafat", Description: "Warm"},
}

// openaiVoices are the voices for OpenAI-compatible TTS models.
var openaiVoices = []contracts.Voice{
	{ID: "alloy", Label: "Alloy", Description: "Neutral, balanced"},
	{ID: "echo", Label: "Echo", Description: "Clear, articulate"},
	{ID: "fable", Label: "Fable", Description: "Warm, expressive"},
	{ID: "onyx", Label: "Onyx", Description: "Deep, authoritative"},
	{ID: "nova", Label: "Nova", Description: "Friendly, upbeat"},
	{ID: "shimmer", Label: "Shimmer", Description: "Soft, gentle"},
}

// ttsProvider reports the active TTS provider, derived from the configured
// model.
func ttsProvider() string {
	if strings.Contains(ttsModel(), "gemini") {
		return "gemini"
	}
	return "openai"
}

func providerVoices(provider string) []contracts.Voice {
	if provider == "gemini" {
		return geminiVoices
	}
	return openaiVoices
}

// defaultVoice returns the configured default (env override) or the active
// provider's built-in default.
func defaultVoice(provider string) string {
	if v := os.Getenv("STUDYAUDIO_DEFAULT_VOICE"); v != "" {
		return v
	}
	if v := os.Getenv("TTS_VOICE"); v != "" {
		return v
	}
	if provider == "gemini" {
		return "Erinome"
	}
	return "alloy"
}

func voiceInSet(voice string, set []contracts.Voice) bool {
	for _, v := range set {
		if strings.EqualFold(v.ID, voice) {
			return true
		}
	}
	return false
}

// voiceID resolves the voice to send to the TTS engine. An empty voice uses
// the provider default. A voice that clearly belongs to the *other* provider
// (e.g. an OpenAI "alloy" requested while Gemini is active) is a mismatch that
// would fail the request, so it falls back to the provider default. An unknown
// voice is trusted (it may be a newer voice the catalog doesn't list yet).
func voiceID(voice string) string {
	provider := ttsProvider()
	voice = strings.TrimSpace(voice)
	if voice == "" {
		return defaultVoice(provider)
	}
	if voiceInSet(voice, providerVoices(provider)) {
		return voice
	}
	other := "openai"
	if provider == "openai" {
		other = "gemini"
	}
	if voiceInSet(voice, providerVoices(other)) {
		return defaultVoice(provider)
	}
	return voice
}

// ListVoices is the handler for the list_voices tool. The Synthesize form
// fetches it so the voice picker always matches the active TTS provider.
func ListVoices(_ context.Context, _ contracts.ListVoicesInput) (tool.Result[contracts.ListVoicesOutput], error) {
	provider := ttsProvider()
	voices := providerVoices(provider)
	def := defaultVoice(provider)
	return tool.Result[contracts.ListVoicesOutput]{
		Text: fmt.Sprintf("%d %s voice(s); default %s", len(voices), provider, def),
		Structured: contracts.ListVoicesOutput{
			Kind:         "voices",
			Provider:     provider,
			Model:        ttsModel(),
			DefaultVoice: def,
			Voices:       voices,
		},
	}, nil
}
