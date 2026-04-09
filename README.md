# UCLAW — Sovereign Engineering Habitat

> **Status:** Prototype scaffold — architecture-first, implementation follows.
> This is not a terminal chatbot. Not a VS Code plugin. Not a better CLI.
> UCLAW is the place an engineer opens first.

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
├── core/                     # Core runtime modules
│   ├── world/                # World state graph
│   ├── agents/               # Agent runtime & orchestration
│   ├── memory/               # Knowledge graph vault
│   ├── artifacts/            # Artifact tracker
│   ├── voice/                # Voice dispatch layer
│   └── policies/             # Policy engine
├── cli/                      # uclaw CLI entry point
├── desktop/                  # Spatial desktop shell
├── integrations/             # External provider adapters
├── verification/             # Verifier teams, test matrices, review queues
├── observability/            # Audit events, budget tracking, health panels
├── scripts/                  # Dev tooling
└── .uclaw/                   # Local env, secrets (gitignored)
```

---

## Quick Start (scaffold)

```bash
git clone https://github.com/stizzfer36-del/UCLAW
cd UCLAW
cp .uclaw/.env.example .uclaw/.env
# fill in provider keys
bash scripts/bootstrap.sh
uclaw
```

---

## Phases

See [`docs/roadmap/PHASES.md`](docs/roadmap/PHASES.md) for the full build-verify-run breakdown.

---

## License

MIT — sovereign, local-first, open.
