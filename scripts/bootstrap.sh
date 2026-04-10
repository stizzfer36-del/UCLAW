#!/usr/bin/env bash
# UCLAW Bootstrap — one-command setup (Claude Code style)
# Usage: scripts/bootstrap.sh

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.."; pwd)"
cd "$ROOT_DIR"

echo "==> UCLAW Bootstrap"

# Check dependencies
command -v go >/dev/null 2>&1 || { echo "ERROR: Go is required — https://go.dev/dl/"; exit 1; }
command -v node >/dev/null 2>&1 || echo "WARNING: Node.js not found — desktop Electron shell will not build"
command -v sqlite3 >/dev/null 2>&1 || echo "WARNING: sqlite3 CLI not found — schema inspection limited"

# Create local UCLAW directories
echo "==> Setting up ~/.uclaw directories"
mkdir -p ~/.uclaw/vault/decisions
mkdir -p ~/.uclaw/vault/prompts
mkdir -p ~/.uclaw/vault/sources
mkdir -p ~/.uclaw/vault/logs
mkdir -p ~/.uclaw/vault/notes
mkdir -p ~/.uclaw/vault/research
mkdir -p ~/.uclaw/checkpoints
mkdir -p ~/.uclaw/agents
mkdir -p ~/.uclaw/audit

if [ ! -f ~/.uclaw/.env ]; then
  if [ -f .uclaw/.env.example ]; then
    cp .uclaw/.env.example ~/.uclaw/.env
    echo "==> Created ~/.uclaw/.env from example — fill in your provider keys"
  else
    cat > ~/.uclaw/.env << 'ENV'
# UCLAW environment — fill in your keys
ANTHROPIC_API_KEY=
OPENAI_API_KEY=
UCLAW_LLM_PROVIDER=anthropic
UCLAW_DEFAULT_MODEL=claude-3-7-sonnet-20250219
ENV
    echo "==> Created blank ~/.uclaw/.env — fill in your provider keys"
  fi
fi

# Install / tidy Go dependencies
echo "==> Installing Go dependencies (GOPROXY=direct)"
export GOPROXY="${GOPROXY:-direct}"
export GOFLAGS="${GOFLAGS:-}"

go get github.com/mattn/go-sqlite3@latest
go get github.com/spf13/cobra@latest
go get gopkg.in/yaml.v3@latest
go mod tidy

# Build the UCLAW binary
echo "==> Building UCLAW CLI binary"
go build -o ./uclaw ./cli

echo "==> Build complete. Binary: $(pwd)/uclaw"

# Optional: add to PATH hint
if ! echo "$PATH" | grep -q "$(pwd)"; then
  echo ""
  echo "    TIP: Add UCLAW to your PATH permanently:"
  echo "    echo 'export PATH=\$PATH:$(pwd)' >> ~/.bashrc && source ~/.bashrc"
  echo ""
fi

# Init world state
echo "==> Initializing UCLAW world state"
./uclaw init || echo "WARNING: init returned non-zero — check output above"

# Show status
echo "==> Checking UCLAW status"
./uclaw status || echo "WARNING: status returned non-zero — check output above"

echo ""
echo "====================================="
echo " UCLAW is ready."
echo "====================================="
echo " Run:  ./uclaw desktop --html"
echo " Or:   ./uclaw agent"
echo " Or:   ./uclaw mission"
echo " Or:   ./uclaw --help"
echo "====================================="
