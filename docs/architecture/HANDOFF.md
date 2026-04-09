# UCLAW Handoff State

This document is the repo-level handoff truth for the code that exists today.
If this file and the code disagree, the code wins and this file should be updated immediately.

## What Exists Now

- Runtime entrypoint: `cmd/uclaw/main.go`
- Command router: `internal/app/app.go`
- Local state: SQLite world DB plus vault graph storage under `~/.uclaw/`
- Audit: append-only JSONL with per-command trace IDs in `internal/audit/`
- Agents: spawn, inspect, pause, resume, retire, approval gating in `internal/agents/`
- Memory: node/edge creation, query, search, markdown vault writes in `internal/memory/`
- Artifacts: creation, verification, sign-off, flag, revert in `internal/artifacts/`
- Missions: planner flow, status gates, review queue, checkpoints, replay, rollback in `internal/missions/`
- Planner: orchestration and default team matrix wiring in `internal/planner/`
- Desktop: local TUI and HTML state renderers plus an Electron app scaffold in `desktop/`
- Voice: transcript/file mission dispatch plus live local capture/STT command execution in `internal/voice/`
- Hardening: policy tightening, recovery, merge-aware peer sync, plugin scaffold, docs generation, secret scan in `internal/hardening/`

## Local State Layout

By default the runtime resolves state under `~/.uclaw/` or `UCLAW_HOME` when set.

- `world.db`: main world and runtime state
- `vault/`: markdown and graph-backed memory store
- `audit.jsonl`: append-only audit stream
- `agents/`: handbooks and agent-local state
- `checkpoints/`: mission checkpoint snapshots
- `missions/`: mission working data
- `desktop/`: generated desktop outputs
- `sync-conflicts/`: conflict records from sync import
- `plugins/`: local plugin scaffolds
- `site/`: generated docs output

## Verification Confidence

Verified locally with:

```bash
GOCACHE=/tmp/uclaw-gocache go test ./...
```

The current test suite covers:

- schema migrations and init flow
- IPC behavior in the sandboxed environment
- agent policy and approval gates
- memory graph round trips
- artifact verification and sign-off rules
- mission gating, replay, checkpoints, review flow
- planner orchestration
- desktop renderer state
- transcript-driven voice dispatch
- hardening flows including sync and secret scan

## Material Phase Status

- Phase 0: mostly complete; real socket bind/connect verification is environment-limited
- Phase 1: materially complete and trustworthy
- Phase 2: materially complete
- Phase 3: materially complete
- Phase 4: materially complete
- Phase 5: materially implemented; Electron scaffold and local desktop assets exist, but bundled binary verification still depends on installing Electron toolchain dependencies
- Phase 6: materially implemented; transcript dispatch and live local capture/STT command path exist, but no default speech model is bundled
- Phase 7: materially implemented; audit, policy, recovery, merge-aware peer sync, plugin scaffold, docs generation, and compliance scaffolding exist, but full metrics export remains incomplete

## Known Non-Goals In This Repo State

These are the main places an infrastructure handoff still needs to stay literal and honest:

- No verified bundled Electron application in this workspace because Electron dependencies are not installed locally
- No bundled default offline speech model; live voice depends on a configured local STT command
- No network transport layer for peer sync beyond peer package exchange
- No full metrics backend or trace export pipeline beyond structured audit events and command-level trace IDs

## Handoff Rule

Do not describe UCLAW as a completed production platform.
Describe it as a local-first, implementation-backed runtime prototype with strong core workflow integrity and explicit remaining last-mile gaps.
