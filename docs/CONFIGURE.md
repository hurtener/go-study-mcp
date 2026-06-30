# Configuration Guide

How to connect go-study-mcp to your AI coding assistant.

## Prerequisites

1. Download the binary for your platform from [Releases](https://github.com/hurtener/go-study-mcp/releases)
2. Get an API key from [OpenRouter](https://openrouter.ai)

## Claude Desktop (macOS)

Edit `~/Library/Application Support/Claude/claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "go-study-mcp": {
      "command": "/Users/you/Downloads/go-study-mcp-macos-arm64",
      "env": {
        "OPENROUTER_API_KEY": "sk-or-v1-your-key-here"
      }
    }
  }
}
```

Restart Claude Desktop. The tools appear automatically.

## Claude Desktop (Windows)

Edit `%APPDATA%\Claude\claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "go-study-mcp": {
      "command": "C:\\Users\\you\\Downloads\\go-study-mcp-windows-amd64.exe",
      "env": {
        "OPENROUTER_API_KEY": "sk-or-v1-your-key-here"
      }
    }
  }
}
```

## Claude Code (CLI)

```bash
claude mcp add go-study-mcp \
  -e OPENROUTER_API_KEY=sk-or-v1-your-key-here \
  -- /Users/you/Downloads/go-study-mcp-macos-arm64
```

Verify it's registered:

```bash
claude mcp list
```

## OpenAI Codex

Add to your Codex MCP configuration (typically `~/.codex/config.json` or via the UI):

```json
{
  "mcpServers": {
    "go-study-mcp": {
      "command": "/Users/you/Downloads/go-study-mcp-macos-arm64",
      "env": {
        "OPENROUTER_API_KEY": "sk-or-v1-your-key-here"
      }
    }
  }
}
```

## Cursor

Add to your Cursor MCP settings (`.cursor/mcp.json` or Settings > MCP):

```json
{
  "mcpServers": {
    "go-study-mcp": {
      "command": "/Users/you/Downloads/go-study-mcp-macos-arm64",
      "env": {
        "OPENROUTER_API_KEY": "sk-or-v1-your-key-here"
      }
    }
  }
}
```

## Windsurf

Add to your Windsurf MCP configuration (`.windsurfrules` or Settings > MCP):

```json
{
  "mcpServers": {
    "go-study-mcp": {
      "command": "/Users/you/Downloads/go-study-mcp-macos-arm64",
      "env": {
        "OPENROUTER_API_KEY": "sk-or-v1-your-key-here"
      }
    }
  }
}
```

## Verifying it works

After connecting, ask your assistant:

> "What MCP tools do you have available?"

You should see:
- `generate_podcast`
- `generate_study_guide`
- `generate_flashcards`
- `synthesize_speech`
- `list_jobs` — list audio generation jobs and their status
- `read_audio` — return a generated file as a data URI for inline playback

Then try:

> "Generate a study guide from this material: [paste your notes]"

Audio generation runs **asynchronously**: the tool returns immediately with a
job handle, and the audio appears in the studio UI's **Jobs** tab when ready
(play it inline, or open the saved file). Long content no longer risks a
tool-call timeout.

## Where audio is saved

The server owns the output directory — callers never choose the path (a
caller-chosen path is unreliable across hosts: read-only working directories,
non-existent container mounts). Files are written to, in order of preference:

1. `STUDYAUDIO_OUTPUT_DIR` if set (a relative value is anchored under your home
   directory, never the host's launch directory),
2. `~/go-study-mcp` otherwise,
3. the OS temp directory as a last resort.

Set an explicit location with `STUDYAUDIO_OUTPUT_DIR`:

```json
{
  "mcpServers": {
    "go-study-mcp": {
      "command": "/path/to/go-study-mcp",
      "env": {
        "OPENROUTER_API_KEY": "sk-or-v1-...",
        "STUDYAUDIO_OUTPUT_DIR": "/Users/you/Music/study-audio"
      }
    }
  }
}
```

## Custom models

Override defaults via environment variables:

```json
{
  "mcpServers": {
    "go-study-mcp": {
      "command": "/path/to/go-study-mcp",
      "env": {
        "OPENROUTER_API_KEY": "sk-or-v1-...",
        "STUDYAUDIO_LLM_MODEL": "anthropic/claude-sonnet-4",
        "STUDYAUDIO_TTS_MODEL": "google/gemini-3.1-flash-tts-preview",
        "STUDYAUDIO_DEFAULT_VOICE": "Erinome"
      }
    }
  }
}
```

## Troubleshooting

**Tools don't appear in the assistant**
- Restart the host application after editing config
- Check the binary path is correct and executable (`chmod +x` on macOS/Linux)

**"No API key found" error**
- Ensure `OPENROUTER_API_KEY` is set in the env block of your config

**Timeout on long content**
- Audio generation is asynchronous: the tool returns a job handle immediately
  and the work continues in the background, so long content no longer blocks the
  call. Track progress in the studio UI's **Jobs** tab, or ask the assistant to
  `list_jobs`.

**Can't find the generated audio**
- Files are saved under `STUDYAUDIO_OUTPUT_DIR` (or `~/go-study-mcp` by default).
  The Jobs tab shows the exact path for each completed job.

**Audio sounds robotic**
- The default TTS model is `tts-1`. For best results, use `google/gemini-3.1-flash-tts-preview` which supports expressive voice tags.
