// Package prompts contains localized prompt templates for LLM-driven content
// generation. Each language has its own set of templates so that framing text,
// instructions, and metadata are fully localized — not just the LLM output.
package prompts

import (
	"fmt"
	"strings"
)

// PodcastParams holds the variables injected into a podcast prompt template.
type PodcastParams struct {
	Language       string
	Tone           string
	Persona        string
	TopicHint      string
	DurationTarget string
	WordTarget     int
	Content        string
}

// FlashcardParams holds the variables injected into a flashcard prompt template.
type FlashcardParams struct {
	Language   string
	Difficulty string
	CardCount  int
	Content    string
}

// PodcastScript returns a full system+user prompt pair for podcast generation.
// The system prompt establishes the narrator persona and rules; the user prompt
// delivers the study material.
func PodcastScript(p PodcastParams) (system string, user string) {
	lang := normalizeLang(p.Language)
	tone := p.Tone
	if tone == "" {
		tone = "casual"
	}
	persona := p.Persona
	if persona == "" {
		persona = defaultPersona(lang, tone)
	}

	system = fmt.Sprintf(podcastSystemTemplate(lang), persona, tone, p.WordTarget)
	user = fmt.Sprintf(podcastUserTemplate(lang), p.TopicHint, p.DurationTarget, p.Content)
	return system, user
}

// FlashcardExtract returns a system+user prompt pair for flashcard extraction.
func FlashcardExtract(p FlashcardParams) (system string, user string) {
	lang := normalizeLang(p.Language)
	difficulty := p.Difficulty
	if difficulty == "" {
		difficulty = "intermediate"
	}

	system = fmt.Sprintf(flashcardSystemTemplate(lang), difficulty)
	user = fmt.Sprintf(flashcardUserTemplate(lang), p.CardCount, p.Content)
	return system, user
}

// ──────────────────────────────────────────────────────────────────────────────
// Templates
// ──────────────────────────────────────────────────────────────────────────────

func podcastSystemTemplate(lang string) string {
	if lang == "es" {
		return `Eres %s un narrador de podcast educativo experimentado. Transforma el material de estudio que el usuario proporciona en un script de narración que un oyente pueda seguir solo por oído, sin ayudas visuales.

Reglas:
- Abre con un gancho breve y atractivo que contextualice el tema.
- Resume en prosa fluida y hablada. Sin viñetas, encabezados ni markdown.
- Usa frases de conexión naturales para audio ("Como acabamos de cubrir...", "Ahora, pasemos a...").
- Incluye al menos una analogía o ejemplo del mundo real por cada concepto importante.
- Cierra con un resumen de los 3 puntos más importantes.
- Objetivo de extensión: %d palabras.
- Escribe íntegramente en español. No mezcles idiomas.

Devuelve solo el script de narración. Sin preámbulo, sin meta-comentarios.`
	}

	return `You are %s, an experienced educational podcast narrator. Transform the study material the user provides into a spoken-word narration script that a listener can follow by ear alone, with no visual aids.

Rules:
- Open with a brief, engaging hook that contextualizes the topic.
- Summarize in flowing, spoken-word prose. No bullet points, headers, markdown.
- Use connective phrases natural to audio ("As we just covered...", "Now, let's move on to...").
- Include at least one analogy or real-world example per major concept.
- Close with a recap of the 3 most important takeaways.
- Target length: %d words.
- Write entirely in English. Do not mix languages.

Output only the narration script itself. No preamble, no meta-commentary.`
}

func podcastUserTemplate(lang string) string {
	if lang == "es" {
		return `Tema: %s
Duración objetivo: %s

Material de estudio:
---
%s
---`
	}

	return `Topic: %s
Duration target: %s

Study material:
---
%s
---`
}

func flashcardSystemTemplate(lang string) string {
	if lang == "es" {
		return `Eres un analista de material de estudio. Extrae pares de preguntas y respuestas estilo flashcard, adecuados para una sesión de estudio con recall activo.

Reglas:
- Genera exactamente la cantidad solicitada de pares pregunta-respuesta.
- Calibra la dificultad:
    básico -> definición/recall ("¿Qué es X?")
    intermedio -> comprensión/comparación
    avanzado -> aplicación/síntesis
- Escribe las preguntas y respuestas en español.
- Mantén las respuestas concisas pero completas (1-3 oraciones).
- Cada respuesta debe ser autocontenida y derivable del material.

Devuelve solo un array JSON de objetos {question, answer}.`
	}

	return `You are a study-material analyst. Extract flashcard-style question-and-answer pairs from the material, suitable for an active-recall study session.

Rules:
- Generate exactly the requested number of Q&A pairs.
- Calibrate difficulty:
    basic -> definition/recall ("What is X?")
    intermediate -> comprehension/comparison
    advanced -> application/synthesis
- Write Q&A in the requested language.
- Keep answers concise but complete (1–3 sentences).
- Each answer must be self-contained and derivable from the material.

Return only a JSON array of {question, answer} objects.`
}

func flashcardUserTemplate(lang string) string {
	if lang == "es" {
		return `Genera %d tarjetas de la siguiente materia de estudio:
---
%s
---`
	}

	return `Generate %d flashcards from the following study material:
---
%s
---`
}

// ──────────────────────────────────────────────────────────────────────────────
// Helpers
// ──────────────────────────────────────────────────────────────────────────────

func normalizeLang(lang string) string {
	lang = strings.TrimSpace(strings.ToLower(lang))
	if lang == "" {
		return "en"
	}
	// Accept full BCP-47 tags but strip region
	if i := strings.Index(lang, "-"); i > 0 {
		lang = lang[:i]
	}
	if i := strings.Index(lang, "_"); i > 0 {
		lang = lang[:i]
	}
	return lang
}

func defaultPersona(lang, tone string) string {
	personas := map[string]map[string]string{
		"en": {
			"casual":       "a friendly tutor",
			"academic":     "a university professor",
			"enthusiastic": "an excited science communicator",
			"calm":         "a calm, methodical narrator",
		},
		"es": {
			"casual":       "un tutor amigable",
			"academic":     "un profesor universitario",
			"enthusiastic": "un comunicador científico entusiasta",
			"calm":         "un narrador calmado y metódico",
		},
	}
	if langMap, ok := personas[lang]; ok {
		if p, ok := langMap[tone]; ok {
			return p
		}
	}
	return "a friendly tutor"
}
