package handlers

import (
	"strings"
	"testing"
)

func TestChunkTextShortReturnsSingle(t *testing.T) {
	got := chunkText("short text", 1500)
	if len(got) != 1 || got[0] != "short text" {
		t.Fatalf("chunkText = %#v, want single chunk", got)
	}
}

func TestChunkTextRespectsLimit(t *testing.T) {
	// 100 sentences of ~20 chars each, chunked at 200.
	var sb strings.Builder
	for i := 0; i < 100; i++ {
		sb.WriteString("This is a sentence. ")
	}
	chunks := chunkText(sb.String(), 200)
	if len(chunks) < 2 {
		t.Fatalf("expected multiple chunks, got %d", len(chunks))
	}
	for i, c := range chunks {
		if n := len([]rune(c)); n > 200 {
			t.Errorf("chunk %d has %d runes, exceeds limit 200", i, n)
		}
		if c == "" {
			t.Errorf("chunk %d is empty", i)
		}
	}
}

func TestChunkTextReassembles(t *testing.T) {
	text := "Alpha beta gamma delta epsilon zeta eta theta iota kappa lambda mu nu xi omicron pi rho sigma tau upsilon."
	chunks := chunkText(text, 30)
	rejoined := strings.Join(chunks, " ")
	// Every original word should survive chunking (no mid-word cuts dropping data).
	for _, w := range strings.Fields(text) {
		if !strings.Contains(rejoined, strings.Trim(w, ".")) {
			t.Errorf("word %q lost during chunking", w)
		}
	}
}

func TestRemovePauseMarkers(t *testing.T) {
	got := removePauseMarkers("Hello [PAUSE:3] world [PAUSE:10] done")
	if strings.Contains(got, "PAUSE") {
		t.Errorf("pause markers not removed: %q", got)
	}
}
