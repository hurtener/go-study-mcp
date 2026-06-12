#!/bin/bash
# generate-screenshots.sh — Run locally to capture inspector screenshots
#
# Usage:
#   ./scripts/generate-screenshots.sh
#
# Prerequisites:
#   - dockyard CLI installed
#   - go-study-mcp built (go build -o go-study-mcp .)
#   - OPENROUTER_API_KEY set in .env
#
# This starts the server, attaches the inspector, and opens your browser.
# Take screenshots manually from the inspector UI.

set -euo pipefail

cd "$(dirname "$0")/.."

DOCKYARD="$(go env GOPATH)/bin/dockyard"

# Build if needed
if [ ! -f go-study-mcp ]; then
  echo "Building go-study-mcp..."
  go build -o go-study-mcp .
fi

# Source env
if [ -f .env ]; then
  set -a
  source .env
  set +a
fi

PORT=${PORT:-8080}

echo "Starting server on port $PORT (HTTP transport)..."
DOCKYARD_TRANSPORT=http DOCKYARD_HTTP_ADDR="127.0.0.1:$PORT" ./go-study-mcp &
SERVER_PID=$!
sleep 2

echo "Starting inspector on port ${INSPECTOR_PORT:-0}..."
$DOCKYARD inspect --url "http://127.0.0.1:$PORT" --dir . &
INSPECTOR_PID=$!

echo ""
echo "═══════════════════════════════════════════════════════════"
echo "  Inspector is running. Look for the URL in the output above."
echo ""
echo "  1. Open the inspector URL in your browser"
echo "  2. Select a fixture from the Fixtures panel"
echo "  3. Click 'Preview' to see the UI"
echo "  4. Take a screenshot"
echo ""
echo "  Fixtures available:"
echo "    - generate_podcast/ready.json"
echo "    - generate_study_guide/ready.json"
echo "    - generate_flashcards/ready.json"
echo "    - synthesize_speech/ready.json"
echo ""
echo "  Press Ctrl+C to stop"
echo "═══════════════════════════════════════════════════════════"
echo ""

cleanup() {
  echo "Stopping..."
  kill $INSPECTOR_PID 2>/dev/null || true
  kill $SERVER_PID 2>/dev/null || true
}
trap cleanup EXIT

wait
