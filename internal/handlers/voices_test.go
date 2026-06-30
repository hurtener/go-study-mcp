package handlers

import (
	"os"
	"testing"
)

func TestVoiceIDValidation(t *testing.T) {
	// Default config → Gemini provider.
	os.Unsetenv("STUDYAUDIO_DEFAULT_VOICE")
	os.Unsetenv("TTS_VOICE")
	os.Unsetenv("STUDYAUDIO_TTS_MODEL")
	os.Unsetenv("TTS_MODEL")

	if got := voiceID(""); got != "Erinome" {
		t.Errorf("empty voice = %q, want default Erinome", got)
	}
	if got := voiceID("Puck"); got != "Puck" {
		t.Errorf("valid Gemini voice = %q, want Puck", got)
	}
	// "alloy" is an OpenAI voice — a cross-provider mismatch on Gemini falls
	// back to the provider default rather than failing the request.
	if got := voiceID("alloy"); got != "Erinome" {
		t.Errorf("cross-provider voice = %q, want fallback Erinome", got)
	}
	// An unknown voice is trusted (may be a newer voice not in the catalog).
	if got := voiceID("Brandnew"); got != "Brandnew" {
		t.Errorf("unknown voice = %q, want passthrough", got)
	}
}

func TestVoiceIDOpenAIProvider(t *testing.T) {
	os.Setenv("STUDYAUDIO_TTS_MODEL", "openai/gpt-4o-mini-tts")
	defer os.Unsetenv("STUDYAUDIO_TTS_MODEL")
	os.Unsetenv("STUDYAUDIO_DEFAULT_VOICE")
	os.Unsetenv("TTS_VOICE")

	if ttsProvider() != "openai" {
		t.Fatalf("provider = %q, want openai", ttsProvider())
	}
	if got := voiceID(""); got != "alloy" {
		t.Errorf("empty voice = %q, want default alloy", got)
	}
	// "Erinome" is a Gemini voice — mismatch on OpenAI falls back.
	if got := voiceID("Erinome"); got != "alloy" {
		t.Errorf("cross-provider voice = %q, want fallback alloy", got)
	}
	if got := voiceID("nova"); got != "nova" {
		t.Errorf("valid OpenAI voice = %q, want nova", got)
	}
}

func TestDefaultVoiceEnvOverride(t *testing.T) {
	os.Setenv("STUDYAUDIO_DEFAULT_VOICE", "Sulafat")
	defer os.Unsetenv("STUDYAUDIO_DEFAULT_VOICE")
	if got := defaultVoice("gemini"); got != "Sulafat" {
		t.Errorf("defaultVoice = %q, want env override Sulafat", got)
	}
}
