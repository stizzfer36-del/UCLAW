#!/usr/bin/env bash
# UCLAW bootstrap — sets up the dev environment.
set -euo pipefail

echo "==> UCLAW bootstrap"

# Create local env directory
mkdir -p .uclaw/vault

# Copy env example if not present
if [ ! -f .uclaw/.env ]; then
  cp .uclaw/.env.example .uclaw/.env 2>/dev/null || true
  echo "  Created .uclaw/.env — fill in your provider keys."
fi

# Check Go
if ! command -v go &>/dev/null; then
  echo "  ERROR: Go not found. Install from https://go.dev/dl/"
  exit 1
fi
echo "  Go: $(go version)"

# Tidy dependencies
go mod tidy 2>/dev/null || echo "  (go mod tidy skipped — no network or missing deps)"

# Build CLI
echo "==> Building uclaw CLI..."
go build -o bin/uclaw ./cli/ 2>/dev/null && echo "  Built: ./bin/uclaw" || echo "  Build skipped (CGo/sqlite3 may need: apt-get install gcc libsqlite3-dev)"

echo ""
echo "==> Done. Next steps:"
echo "  1. Edit .uclaw/.env with your API keys"
echo "  2. ./bin/uclaw daemon   # start the runtime"
echo "  3. ./bin/uclaw --help   # explore commands"
