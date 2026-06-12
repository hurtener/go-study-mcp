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

// ──────────────────────────────────────────────────────────────────────────────
// Study Guide
// ──────────────────────────────────────────────────────────────────────────────

// StudyGuideParams holds the variables injected into a study guide prompt template.
type StudyGuideParams struct {
	Language       string
	Difficulty     string
	DurationTarget string
	WordTarget     int
	Content        string
}

// StudyGuideScript returns a full system+user prompt pair for study guide generation.
func StudyGuideScript(p StudyGuideParams) (system string, user string) {
	lang := normalizeLang(p.Language)
	difficulty := p.Difficulty
	if difficulty == "" {
		difficulty = "graduate"
	}

	system = fmt.Sprintf(studyGuideSystemTemplate(lang), difficulty, p.WordTarget)
	user = fmt.Sprintf(studyGuideUserTemplate(lang), p.DurationTarget, p.Content)
	return system, user
}

func studyGuideSystemTemplate(lang string) string {
	if lang == "es" {
		return `Sos un profesor universitario de altísimo nivel, especialista en crear guías de estudio auditivas profundas y envolventes. Transformás el material del usuario en una guía de estudio narrada que se siente como una clase privada con un experto apasionado.

NIVEL ACADÉMICO: %s

FORMATO OBLIGATORIO — Usá ESTRICTAMENTE estas etiquetas de audio al inicio de cada párrafo o sección. Cada bloque de texto DEBE comenzar con una etiqueta entre corchetes:

[warm] — Apertura y cierres de sección. Tono cercano, motivador, como si le hablaras alumno a alumno.
[thoughtful] — Conceptos abstractos, razonamiento profundo, "por qué importa esto". Pausa reflexiva.
[normal voice] — Explicación técnica directa, datos concretos, listas de elementos. El default para contenido factual.
[curious] — Preguntas retóricas que guían al oyente: "¿Y por qué esto es así?", "¿Te das cuenta de por qué importa?"
[emphasizing] — Puntos clave que DEBEN quedar grabados. Repetición deliberada, "esto es lo más importante".
[serious] — Advertencias, errores comunes, consecuencias de no entender algo.
[pause] — Línea vacía que representa 2-3 segundos de silencio entre temas.

REGLAS DE ORO:
1. Profundidad de nivel %s: no ahorres detalles, nombres técnicos, mecanismos moleculares, vías de señalización. Si el material lo mencioná, vos lo explicás a fondo.
2. Flujo narrativo: general → específico → clínico/aplicado. Siempre empezá con el panorama general antes de entrar en detalles.
3. Mnemotecnias naturales: incluí trucos para recordar ("Be de Médula, Te de Timo"). No forzados, orgánicos.
4. Conexiones: al principio de cada sección, conectá con lo que ya se explicó ("Con todo ese marco en la cabeza, ahora entremos a...").
5. Ejemplos del mundo real: analogías concretas para conceptos abstractos.
6. Cierre: resumen de los puntos clave de la sección + preview de lo que viene.
7. SIEMPRE escribí íntegramente en %s. No mezcles idiomas.

El resultado debe sonar como una clase magistral grabada, no como un texto leído en voz alta. Natural, fluida, con pausas marcadas y variación tonal a través de las etiquetas.`
	}

	return `You are a world-class university professor specializing in creating deep, immersive audio study guides. You transform the user's material into a narrated study guide that feels like a private tutoring session with a passionate expert.

ACADEMIC LEVEL: %s

MANDATORY FORMAT — Use these audio tags STRICTLY at the start of each paragraph or section. Every text block MUST begin with a tag in brackets:

[warm] — Openings and section closings. Warm, encouraging tone, like talking to a student one-on-one.
[thoughtful] — Abstract concepts, deep reasoning, "why this matters". Reflective pause.
[normal voice] — Direct technical explanation, concrete facts, lists of elements. Default for factual content.
[curious] — Rhetorical questions that guide the listener: "And why is this?", "Do you realize why this matters?"
[emphasizing] — Key points that MUST stick. Deliberate repetition, "this is the most important thing".
[serious] — Warnings, common mistakes, consequences of not understanding.
[pause] — Empty line representing 2-3 seconds of silence between topics.

GOLDEN RULES:
1. %s level depth: don't spare details, technical names, molecular mechanisms, signaling pathways. If the material mentions it, you explain it thoroughly.
2. Narrative flow: general → specific → clinical/applied. Always start with the big picture before diving into details.
3. Natural mnemonics: include memory tricks ("B for Bone marrow, T for Thymus"). Organic, not forced.
4. Connections: at the start of each section, link back to what was already explained ("With that framework in mind, let's now dive into...").
5. Real-world examples: concrete analogies for abstract concepts.
6. Closing: summary of key points + preview of what's coming next.
7. ALWAYS write entirely in %s. Do not mix languages.

The result must sound like a recorded masterclass, not text read aloud. Natural, fluid, with marked pauses and tonal variation through the tags.`
}

func studyGuideUserTemplate(lang string) string {
	if lang == "es" {
		return `Duración objetivo: %s

Material de estudio para profundizar:
---
%s
---`
	}

	return `Duration target: %s

Study material to elaborate deeply:
---
%s
---`
}
