# go-study-mcp

[![Release](https://img.shields.io/github/v/release/hurtener/go-study-mcp)](https://github.com/hurtener/go-study-mcp/releases)
[![License](https://img.shields.io/github/license/hurtener/go-study-mcp)](https://github.com/hurtener/go-study-mcp/blob/main/LICENSE)
[![Go](https://img.shields.io/badge/go-1.24+-00ADD8?logo=go)](https://go.dev)
[![Svelte](https://img.shields.io/badge/svelte-5-FF3E00?logo=svelte)](https://svelte.dev)

**MCP server that turns study material into narrated audio — podcasts, deep study guides, flashcards, and speech.**

Built with [Dockyard](https://github.com/hurtener/dockyard) · Ships as a single CGo-free binary with an inline Svelte UI.

## What it does

You give it study material. It gives you audio that teaches.

| Tool | What it makes | Mode |
|------|--------------|------|
| `generate_study_guide` | Expert-level narrated guide with expressive voice tags (`[warm]`, `[thoughtful]`, `[curious]`) | Audio + text |
| `generate_podcast` | Casual or academic podcast narration from your notes | Audio + text |
| `generate_flashcards` | Q&A with timed pauses for active recall — the audio waits for you to think | Audio + text |
| `synthesize_speech` | Direct TTS with `[PAUSE:N]` markers for custom pacing | Audio |

All tools support **6 languages** (en, es, fr, de, pt, it) and output **MP3 audio** via Bifrost + OpenRouter.

## Quick Start

### Get a binary

```bash
# macOS (Apple Silicon)
curl -sL https://github.com/hurtener/go-study-mcp/releases/latest/download/go-study-mcp-macos-arm64 -o go-study-mcp
chmod +x go-study-mcp

# Linux (amd64)
curl -sL https://github.com/hurtener/go-study-mcp/releases/latest/download/go-study-mcp-linux-amd64 -o go-study-mcp
chmod +x go-study-mcp
```

### Connect to your AI assistant

**Claude Desktop** — add to `~/Library/Application Support/Claude/claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "go-study-mcp": {
      "command": "/path/to/go-study-mcp",
      "env": {
        "OPENROUTER_API_KEY": "sk-or-v1-your-key"
      }
    }
  }
}
```

**Claude Code** — one command:

```bash
claude mcp add go-study-mcp \
  -e OPENROUTER_API_KEY=sk-or-v1-your-key \
  -- /path/to/go-study-mcp
```

**Cursor / Windsurf / Codex** — same pattern, see [docs/CONFIGURE.md](docs/CONFIGURE.md) for all platforms.

### Use it

Just ask your assistant:

> "Generate a master's-level study guide on the cutaneous immune system in Spanish"
> "Turn these notes into a podcast about mitochondria"
> "Create 10 flashcards from this chapter"

## The study guide is the star

Most audio study tools read your notes back to you. This one **teaches**.

The `generate_study_guide` tool produces narrated guides with voice expression tags that Gemini TTS interprets for genuine tonal variation:

```
[warm] Bienvenido a esta guía. Hoy vamos a hablar de algo que llevás encima las 24 horas del día.

[thoughtful] Antes de hablar de la piel en sí, necesitamos tener claro el panorama general de la inmunidad.

[normal voice] La inmunidad innata es la respuesta rápida. Pensala como la primera línea.

[curious] ¿Y por qué primero la innata y después la adaptativa? Porque la innata es veloz.

[emphasizing] La adaptativa deja células de memoria. Esto es lo más importante.
```

Four academic levels: `undergraduate` → `graduate` → `masters` → `phd`.

## Architecture

```
┌─────────────────────────────────────────────┐
│  MCP Host (Claude, Cursor, Codex, etc.)     │
└──────────────────┬──────────────────────────┘
                   │ stdio
┌──────────────────▼──────────────────────────┐
│  go-study-mcp (this binary)                 │
│  ┌─────────────┐  ┌──────────────────────┐  │
│  │ 4 MCP tools │  │ Svelte UI (embedded) │  │
│  └──────┬──────┘  └──────────────────────┘  │
│         │                                    │
│  ┌──────▼──────┐  ┌──────────────────────┐  │
│  │ Bifrost SDK │  │ PCM→MP3 (shine-mp3)  │  │
│  └──────┬──────┘  └──────────────────────┘  │
└─────────┼───────────────────────────────────┘
          │ HTTP
┌─────────▼───────────────────────────────────┐
│  OpenRouter (LLM + TTS)                     │
│  LLM: anthropic/claude-sonnet-4.6           │
│  TTS: google/gemini-3.1-flash-tts-preview   │
└─────────────────────────────────────────────┘
```

## Configuration

| Variable | Description | Default |
|----------|-------------|---------|
| `OPENROUTER_API_KEY` | **Required.** OpenRouter API key | — |
| `STUDYAUDIO_LLM_MODEL` | LLM for text generation | `openai/gpt-4o-mini` |
| `STUDYAUDIO_TTS_MODEL` | TTS model | `tts-1` |
| `STUDYAUDIO_DEFAULT_VOICE` | TTS voice | `alloy` |

## Building from source

```bash
git clone https://github.com/hurtener/go-study-mcp.git
cd go-study-mcp

# Build UI + binary
cd web && npm ci && npm run build && cd ..
go build -o go-study-mcp .

# Run (stdio transport, for MCP hosts)
OPENROUTER_API_KEY=sk-or-v1-... ./go-study-mcp

# Run (HTTP transport, for inspector/debugging)
DOCKYARD_TRANSPORT=http ./go-study-mcp
```

## License

[Apache 2.0](LICENSE)
