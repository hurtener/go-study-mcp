# go-study-mcp

**Turn your notes into study audio.** An MCP server that transforms study material into narrated podcasts, deep study guides, flashcards, and speech — with expressive TTS that actually sounds good.

## What it does

You give it study material. It gives you audio.

| Tool | What it makes |
|------|--------------|
| `generate_podcast` | Narrated podcast from your notes — casual, academic, or custom persona |
| `generate_study_guide` | Deep expert-level study guide with expressive voice tags (`[warm]`, `[thoughtful]`, `[curious]`, `[emphasizing]`) |
| `generate_flashcards` | Q&A flashcards with timed pauses for active recall |
| `synthesize_speech` | Direct text-to-speech with `[PAUSE:N]` markers |

All tools support **multiple languages** (en, es, fr, de, pt, it) and output **MP3 audio**.

## The study guide is the star

Most audio study tools read your notes back to you. This one **teaches**.

The `generate_study_guide` tool produces narrated guides with voice expression tags that Gemini TTS interprets for genuine tonal variation:

```
[warm] Bienvenido a esta guía. Hoy vamos a hablar de algo que llevás encima las 24 horas del día.

[thoughtful] Antes de hablar de la piel en sí, necesitamos tener claro el panorama general de la inmunidad.

[normal voice] La inmunidad innata es la respuesta rápida. Pensala como la primera línea.

[curious] ¿Y por qué primero la innata y después la adaptativa?

[emphasizing] La adaptativa deja células de memoria. Esto es lo más importante.
```

Four academic levels: `undergraduate` → `graduate` → `masters` → `phd`.

## Quick start

### 1. Get an API key

Sign up at [OpenRouter](https://openrouter.ai) and create an API key.

### 2. Configure your host

**Claude Desktop** — edit `~/Library/Application Support/Claude/claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "go-study-mcp": {
      "command": "/path/to/go-study-mcp",
      "env": {
        "OPENROUTER_API_KEY": "sk-or-v1-your-key-here"
      }
    }
  }
}
```

**Claude Code** — run:

```bash
claude mcp add go-study-mcp \
  -e OPENROUTER_API_KEY=sk-or-v1-your-key-here \
  -- /path/to/go-study-mcp
```

**OpenAI Codex** — add to your config:

```json
{
  "mcpServers": {
    "go-study-mcp": {
      "command": "/path/to/go-study-mcp",
      "env": {
        "OPENROUTER_API_KEY": "sk-or-v1-your-key-here"
      }
    }
  }
}
```

### 3. Use it

Just ask your AI assistant to create study audio from any material:

> "Turn these notes into a study podcast about mitochondria"
> "Generate a master's-level study guide on the cutaneous immune system in Spanish"
> "Create flashcards from this chapter with 10 questions"

## Building from source

```bash
git clone https://github.com/hurtener/go-study-mcp.git
cd go-study-mcp

# Build UI
cd web && npm ci && npm run build && cd ..

# Build binary
go build -o go-study-mcp .

# Run
OPENROUTER_API_KEY=sk-or-v1-your-key ./go-study-mcp
```

## Configuration

| Variable | Description | Default |
|----------|-------------|---------|
| `OPENROUTER_API_KEY` | OpenRouter API key (required) | — |
| `STUDYAUDIO_LLM_MODEL` | LLM model for text generation | `openai/gpt-4o-mini` |
| `STUDYAUDIO_TTS_MODEL` | TTS model for audio synthesis | `tts-1` |
| `STUDYAUDIO_DEFAULT_VOICE` | Default TTS voice | `alloy` |

## License

Apache 2.0
