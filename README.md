# go-study-mcp

MCP server to generate study audio content. Built with [Dockyard](https://github.com/hurtener/dockyard).

## Features

- **Podcast generation** — Transform study material into narrated podcast-style audio
- **Flashcard audio** — Generate Q&A flashcard sessions with timed pauses for active recall
- **Speech synthesis** — Direct text-to-speech with pause marker support
- **Multi-language** — Full localization support (English, Spanish, and extensible)
- **Flexible prompting** — Customizable tone, persona, and style

## Tools

| Tool | Description |
|------|-------------|
| `generate_podcast` | Transform study material into narrated audio with flexible tone/persona |
| `generate_flashcards` | Generate Q&A flashcard audio with timed pauses |
| `synthesize_speech` | Direct text-to-speech with [PAUSE:N] markers |

## Development

```bash
# Install dependencies
go mod tidy

# Run tests
go test ./...

# Validate contracts
dockyard validate

# Run the server
go run .
```

## License

Apache 2.0 — see [LICENSE](LICENSE)
