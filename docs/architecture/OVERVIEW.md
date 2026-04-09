# UCLAW Architecture Overview

## Guiding Principle

The habitat owns the workflow, not the vendors.
Every external provider (Anthropic, OpenRouter, local model, CAD tool) plugs in behind a capability layer.
The user's local state graph is the single source of truth.

This document describes the architecture implemented in the current repository, not an aspirational future layout.

---

## Five Integrated Subsystems

### 1. CLI Surface (`cmd/uclaw/`, `internal/app/`)
- Entry point: `uclaw [command]`
- Scripting surface for all habitat operations
- Launches desktop shell or runs headless agent jobs
- Connects to the shared world state graph over a local IPC/socket and direct local runtime calls

### 2. Spatial Desktop Shell (`desktop/`, `internal/observability/`)
- Three-panel layout:
  - **Left sidebar:** Offices → Teams → Members tree
  - **Center canvas:** Up to 6 live panes (terminal, code, doc, CAD, browser, notebook)
  - **Right rail:** Agent roster, workflow queue, status, health
- Built today as a shared-state local HTML renderer plus terminal-first fallback (TUI), with an Electron project scaffold for packaging
- Pane types: terminal, markdown editor, code editor, browser iframe, CAD viewer, diff view

### 3. Multi-Agent Runtime (`internal/agents/`, `internal/planner/`)
- Lead agent + sub-agent hierarchy
- Team orchestration: Planner → Dev Team → Verifier Team
- Each agent is a first-class identity with:
  - Own terminal session
  - Private engineering handbook (reasoning style, citation rules, log format)
  - Allowed capability set (tools, paths, providers)
  - Audit event emitter
- Supports: Claude (Anthropic), OpenRouter models, local (Ollama/LM Studio), custom

### 4. Memory Vault (`internal/memory/`)
- Local-first markdown + frontmatter knowledge graph
- Nodes: decisions, prompts, source links, agent logs, meeting notes
- Edges: `caused-by`, `verified-by`, `supersedes`, `cites`
- Obsidian-compatible folder structure (can open in Obsidian as a side window)
- Graph query API used by agents and desktop canvas

### 5. Artifact System (`internal/artifacts/`, `internal/missions/`)
- Tracks: code files, docs, slides, CAD specs, test reports, screenshots
- Each artifact has: origin agent, source citations, verification status, sign-off chain
- Unprovenanced artifacts flagged as lower trust (visible in UI)
- Git-integrated: artifact records map to commits/branches

## Current Late-Phase Reality

- Observability is implemented through status, health, budget, workflow, error, and timeline loaders backed by the same local state as the CLI.
- Voice is implemented as transcript/file dispatch plus live local capture/STT command execution.
- Hardening includes policy tightening, checkpoint-based recovery, merge-aware peer sync with conflict artifacts, plugin scaffolding, docs generation, and secret scanning.
- The repo includes an Electron scaffold, but this workspace has not installed Electron dependencies, so a bundled app is not locally verified here.

---

## Integration Architecture Diagram

```
┌─────────────────────────────────────────────────┐
│                   UCLAW Habitat                  │
│                                                 │
│  ┌──────────┐  ┌──────────────┐  ┌───────────┐ │
│  │   CLI    │  │ Desktop Shell│  │  Voice    │ │
│  └────┬─────┘  └──────┬───────┘  └─────┬─────┘ │
│       └───────────────┼────────────────┘        │
│                       ▼                         │
│            ┌──────────────────┐                 │
│            │  World State IPC │                 │
│            └────────┬─────────┘                 │
│    ┌────────────────┼────────────────┐           │
│    ▼                ▼               ▼           │
│ ┌──────┐      ┌─────────┐     ┌──────────┐      │
│ │Memory│      │ Agents  │     │Artifacts │      │
│ │Vault │      │Runtime  │     │ Tracker  │      │
│ └──────┘      └────┬────┘     └──────────┘      │
│                    │                            │
│              ┌─────▼──────┐                     │
│              │ Capability │                     │
│              │   Layer    │                     │
│              └─────┬──────┘                     │
└────────────────────┼────────────────────────────┘
                     │
     ┌───────────────┼───────────────┐
     ▼               ▼               ▼
  Anthropic     OpenRouter      Local Models
  (Claude)       (multi)        (Ollama etc)
```
