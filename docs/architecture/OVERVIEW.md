# UCLAW Architecture Overview

## Philosophy

UCLAW is not a chatbot wrapper. It is a **sovereign engineering habitat**:
a local-first OS layer that owns world-state, missions, agents, memory,
artifacts, and policies — and treats every external framework as a
pluggable *capability*, not a foundation.

## Layered Model

```
┌─────────────────────────────────────────────────────────┐
│  CLI  (uclaw)              ← canonical control surface   │
├─────────────────────────────────────────────────────────┤
│  IPC Server  (.uclaw/uclaw.sock)   ← daemon interface    │
├─────────────────────────────────────────────────────────┤
│  Core Runtime                                            │
│    world/     — SQLite world graph                       │
│    agents/    — agent registry + policy gate             │
│    memory/    — FTS5 knowledge vault                     │
│    artifacts/ — artifact tracker + trust levels          │
│    observability/ — chained audit log                    │
├─────────────────────────────────────────────────────────┤
│  Capability Layer  (core/integrations/)                  │
│    fabric/          — prompt/pattern workflows           │
│    codex/           — local code editing (OpenAI Codex)  │
│    local_llm/       — Ollama / LM Studio providers       │
│    voltagent/       — TS multi-agent platform            │
│    praisonai/       — Python multi-agent framework       │
│    agent_framework/ — Microsoft AutoGen orchestration    │
│    observer/        — local model orchestrator           │
│    ros2/            — robotics middleware (ROS2)         │
│    iot_mqtt/        — IoT device control (MQTT)          │
│    doc_parser/      — PDF/DOCX/MD ingestion              │
│    cad/             — FreeCAD headless inspection        │
├─────────────────────────────────────────────────────────┤
│  Policy Engine  (core/policies/tools.yaml +              │
│                  core/policies/integrations.yaml)        │
│    risk_level, requires_approval, whitelist, role gate   │
├─────────────────────────────────────────────────────────┤
│  Verification System  (verification/)                    │
│    checklist.yaml — trust tier checks                    │
│    review_queue.go — pending/approve/reject              │
└─────────────────────────────────────────────────────────┘
```

## World Model

```
World → Offices → Teams → Members → Machines → Rooms
     → Missions → Memory → Artifacts → Capabilities → Policies
```

All external OSS plugs in **under** this model — never beside it.

## Integration Contract

Every integration must:
1. Live under `core/integrations/<name>/`.
2. Expose a Go adapter (`adapter.go`) with typed, error-returning functions.
3. Register its tools in `core/policies/integrations.yaml`.
4. Have every tool call pass through `agents.CheckTool()` before executing.
5. Emit an `observability.Audit` event on every call (pending implementation).

## Key Invariants

- **No external framework owns world state.** Fabric, AutoGen, VoltAgent, etc.
  are all tools. UCLAW's SQLite DB is the single source of truth.
- **Every tool call is policy-gated.** No integration bypasses `CheckTool()`.
- **Audit log is append-only and chained** (SHA-256 prev-hash chain).
- **Artifacts have provenance.** Trust levels: `provenanced` → `partially-provenanced` → `unprovenanced`.
