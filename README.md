# UCLAW — Sovereign Engineering Habitat

> **Status:** Local-first runtime prototype with working core systems.
> The repository is no longer just an architecture scaffold: the Go runtime, SQLite world state, audit log, mission flow, artifact verification, memory vault, planner, desktop renderers, and transcript-driven voice dispatch are implemented and tested.
> It is still not literally roadmap-complete in the last-mile areas called out in `docs/roadmap/PHASES.md`.

---

## What UCLAW Is

UCLAW is a **local-first operating system for engineers** that unifies:

| Layer | What it replaces / unifies |
|---|---|
| CLI surface | All scattered shell sessions |
| Spatial IDE canvas | Editor + browser + terminal tabs |
| Multi-agent runtime | Claude Code + orchestration glue |
| Knowledge graph vault | Obsidian + Notion + scattered notes |
| Artifact system | GitHub + Figma + doc folders |
| Voice layer | Spoken workflow dispatch |

Opening your laptop and typing `uclaw` gives you a **living control room** — not a sidecar.

---

## World Model

```
World
└── Offices
    └── Teams
        └── Members
            └── Machines
                └── Rooms
                    └── Missions
                        └── Memory
                        └── Artifacts
                        └── Capabilities
                        └── Policies
```

---

## Repository Layout

```
UCLAW/
├── docs/                     # Architecture, decisions, roadmap
│   ├── architecture/
│   ├── verification/
│   ├── security/
│   └── roadmap/
├── cmd/uclaw/                # CLI entry point
├── internal/                 # Runtime implementation
│   ├── app/                  # Command routing and orchestration
│   ├── world/                # World state graph and migrations
│   ├── agents/               # Agent runtime, approvals, handbooks
│   ├── memory/               # Knowledge graph vault
│   ├── artifacts/            # Artifact tracker and verification
│   ├── missions/             # Mission flow, checkpoints, review
│   ├── planner/              # Planner and team orchestration
│   ├── observability/        # Status, health, budget, desktop state
│   ├── hardening/            # Recovery, policy, sync, plugins, docs
│   ├── voice/                # Transcript-driven dispatch layer
│   └── audit/                # Structured append-only audit stream
├── core/                     # Shared policy/schema assets
├── desktop/                  # Local HTML/TUI desktop shell assets
├── scripts/                  # Dev tooling
└── .uclaw/                   # Local env, secrets (gitignored)
```

## Current Capability

- `uclaw init` creates the local layout, world database, and audit stream.
- `uclaw agent`, `memory`, `artifact`, `mission`, `plan`, `review`, `override`, `status`, `health`, `budget`, `policy`, `desktop`, `voice`, `sync`, `plugin`, and `docs` all route through the Go runtime today.
- The desktop surface now includes a real Electron project scaffold under `desktop/electron/`, plus the existing shared-state HTML/TUI renderer.
- Voice supports transcript/file dispatch and a live local capture/STT command path.
- Sync now exports and imports peer packages, serves/pulls/pushes over HTTP, applies canonical merge rules by file class, and keeps database files conflict-oriented.

## Quick Start

```bash
git clone https://github.com/stizzfer36-del/UCLAW
cd UCLAW
GOCACHE=/tmp/uclaw-gocache go test ./...
go run ./cmd/uclaw init
go run ./cmd/uclaw status
go run ./cmd/uclaw desktop --html
```

## Phases

See [`docs/roadmap/PHASES.md`](docs/roadmap/PHASES.md) for the full build-verify-run breakdown.

For the current code-backed status and handoff caveats, start with:

- [`docs/architecture/HANDOFF.md`](docs/architecture/HANDOFF.md)
- [`docs/architecture/OVERVIEW.md`](docs/architecture/OVERVIEW.md)
- [`docs/roadmap/PHASES.md`](docs/roadmap/PHASES.md)

---

## License

MIT — sovereign, local-first, open.
