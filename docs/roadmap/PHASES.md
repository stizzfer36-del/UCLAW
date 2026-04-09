# UCLAW Build Phases

## Phase 0 — Scaffold ✅
- Repo structure, README, schema.sql, tools.yaml, .gitignore

## Phase 1 — Base Runtime ✅ (this commit)
- Go core: world/agents/memory/artifacts/ipc/observability
- CLI: `uclaw daemon | world | agent | memory | artifact | audit`
- Policy engine: CheckTool() with tools.yaml + integrations.yaml
- All integration adapters (fabric, codex, local_llm, voltagent, praisonai,
  agent_framework, observer, ros2, iot_mqtt, doc_parser, cad)
- Verification system: checklist.yaml + review_queue.go
- Architecture docs

## Phase 2 — Wired Runtime (next)
- Real agent loop: provider calls (Anthropic / OpenAI / Ollama)
- Mission state machine: active → blocked → complete
- Memory auto-write on every agent turn
- Artifact auto-register on every tool output
- IPC round-trip from CLI → daemon → handler

## Phase 3 — Observability & Verification
- Audit event emission on every tool call
- Hash-chain integrity check command
- Verifier agent role: auto-run checklist.yaml checks
- Budget tracker: token / cost per agent per mission

## Phase 4 — Hardware & Voice
- ROS2 topic whitelist enforcement
- IoT device registry in world schema
- Voice dispatch layer (whisper → intent → uclaw IPC)

## Phase 5 — Desktop Shell
- Spatial canvas (desktop/) — Tauri or Electron wrapper over the CLI daemon
