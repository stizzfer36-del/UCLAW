#!/usr/bin/env bash
# UCLAW Bootstrap Script
# Sets up the local development environment

set -euo pipefail

echo "==> UCLAW Bootstrap"

# Check dependencies
command -v go >/dev/null 2>&1 || { echo "ERROR: Go is required (https://go.dev/dl/)"; exit 1; }
command -v node >/dev/null 2>&1 || { echo "WARNING: Node.js not found — desktop shell will not build"; }
command -v sqlite3 >/dev/null 2>&1 || { echo "WARNING: sqlite3 CLI not found — schema inspection limited"; }

echo "==> Checking .uclaw directory"
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
  cp .uclaw/.env.example ~/.uclaw/.env
  echo "==> Created ~/.uclaw/.env from example — fill in your provider keys"
fi

echo "==> Building CLI (stub)"
# go build -o ~/.local/bin/uclaw ./cli/  # Uncomment when cli/ is implemented

echo "==> Bootstrap complete."
echo "    Next: fill in ~/.uclaw/.env, then run: uclaw init"
