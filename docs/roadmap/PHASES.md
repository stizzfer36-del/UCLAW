# UCLAW Build Roadmap

Every phase has: **Build turns**, **Verification turns**, and **Run gates** before the next phase unlocks.

## Current Repo Status (2026-04-09)

The checklist below is still the roadmap contract, not a claim that every box is literally complete in this environment.
The code in this repository currently stands at:

- Phase 0: mostly complete, with the exact IPC socket verification limited by sandbox constraints
- Phase 1: materially complete
- Phase 2: materially complete
- Phase 3: materially complete
- Phase 4: materially complete
- Phase 5: materially implemented; shared-state HTML/TUI shell and Electron scaffold exist, but local bundle verification depends on Electron dependencies being installed
- Phase 6: materially implemented; transcript/file-driven voice dispatch and live local capture/STT command path exist, but no default speech model is bundled
- Phase 7: materially implemented; audit, policy, recovery, merge-aware peer sync, plugin scaffold, docs generation, and compliance scaffolding exist, but full observability stack proof remains incomplete

Use `docs/architecture/HANDOFF.md` as the implementation-backed summary when handing off the repository.

---

## Phase 0 — Foundation (Weeks 1–3)

### Build Turns
- [ ] B0.1: Project scaffold, repo structure, `go.mod` / `package.json` bootstrapped
- [ ] B0.2: World state SQLite schema + migration scripts
- [ ] B0.3: IPC socket server (world state read/write)
- [ ] B0.4: `uclaw` CLI entry point (cobra/click) with `init`, `world`, `agent` subcommands
- [ ] B0.5: `.uclaw/` local env structure + secrets loader

### Verification Turns
- [ ] V0.1: Schema migration tests (up/down clean)
- [ ] V0.2: IPC socket connection test
- [ ] V0.3: CLI `uclaw init` idempotency test
- [ ] V0.4: Secrets loader — confirm no secrets leak to stdout or logs

### Run Gate
> `uclaw init` creates world.db, prints world ID, no errors. Secrets file present and unlogged.

---

## Phase 1 — Agent Runtime (Weeks 4–7)

### Build Turns
- [ ] B1.1: Agent identity model (ID, name, role, team, capabilities, path whitelist)
- [ ] B1.2: Engineering handbook loader (markdown parse → struct)
- [ ] B1.3: Provider adapter interface + Anthropic Claude adapter
- [ ] B1.4: Tool registry (`tools.yaml` loader, risk levels, approval gating)
- [ ] B1.5: Agent spawn + retire lifecycle
- [ ] B1.6: Audit event emitter (structured JSONL writer)
- [ ] B1.7: OpenRouter adapter
- [ ] B1.8: Local model adapter (Ollama API)

### Verification Turns
- [ ] V1.1: Agent spawn with invalid capability → rejected
- [ ] V1.2: Tool call above risk threshold → approval gate triggered
- [ ] V1.3: Path outside whitelist → rejected, audit event emitted
- [ ] V1.4: Audit log integrity check (no gaps, no tampering)
- [ ] V1.5: Provider adapter mock tests (no real API calls in CI)

### Run Gate
> Spawn agent, assign handbook, call a low-risk tool, verify audit log entry. High-risk tool call blocked until approved.

---

## Phase 2 — Memory Vault (Weeks 8–10)

### Build Turns
- [ ] B2.1: Vault folder structure + frontmatter schema
- [ ] B2.2: Graph DB (SQLite nodes + edges)
- [ ] B2.3: Vault write API (create node, add edge)
- [ ] B2.4: Vault query API (by type, by edge, full-text)
- [ ] B2.5: Agent log writer (missions → vault logs)
- [ ] B2.6: `uclaw memory` CLI commands

### Verification Turns
- [ ] V2.1: Node create/query round-trip test
- [ ] V2.2: Edge traversal depth test
- [ ] V2.3: Full-text search accuracy test
- [ ] V2.4: Obsidian compatibility check (open vault in Obsidian, no corruption)

### Run Gate
> Agent writes a decision node to vault, adds `caused-by` edge, query returns it. Vault readable in Obsidian.

---

## Phase 3 — Artifact System (Weeks 11–13)

### Build Turns
- [ ] B3.1: Artifact record schema + SQLite storage
- [ ] B3.2: Artifact creation on agent file write
- [ ] B3.3: Citation tracker (source URLs, vault links)
- [ ] B3.4: Trust level calculator (provenance score)
- [ ] B3.5: Verification check runner (test, citation, policy)
- [ ] B3.6: Sign-off chain logic
- [ ] B3.7: `uclaw artifact` CLI commands
- [ ] B3.8: Git reference linker (commit SHA → artifact)

### Verification Turns
- [ ] V3.1: Artifact created with no sources → `unprovenanced` flag correct
- [ ] V3.2: Citation completeness ≥80% → `provenanced` flag correct
- [ ] V3.3: Sign-off by builder agent blocked (builder ≠ verifier rule)
- [ ] V3.4: Artifact revert restores previous git SHA cleanly

### Run Gate
> Agent produces a code artifact, verifier team runs checks, sign-off requires separate verifier identity. Unprovenanced artifact visibly flagged.

---

## Phase 4 — Verification Layer (Weeks 14–16)

### Build Turns
- [ ] B4.1: Planner agent team orchestration (lead → dev + verifier spawn)
- [ ] B4.2: Mission test matrix loader (`test_matrix.yaml`)
- [ ] B4.3: Verification gate logic (block mission on check failure)
- [ ] B4.4: Review queue (blocked artifacts visible in CLI + desktop)
- [ ] B4.5: Checkpoint system (save/restore/diff/replay)
- [ ] B4.6: Incident timeline builder
- [ ] B4.7: Handbook amendment generator (from failed verifications)

### Verification Turns
- [ ] V4.1: Mission blocked when citation check fails
- [ ] V4.2: Checkpoint save → restore → state identical (round-trip)
- [ ] V4.3: Mission replay from checkpoint produces same artifact
- [ ] V4.4: Handbook amendment generated after 3 citation failures by same agent
- [ ] V4.5: Builder ≠ Verifier enforced at team level

### Run Gate
> Full mission: planner → dev team builds → verifier team checks → one check fails → mission blocked → engineer reviews → check passes → sign-off → checkpoint saved.

---

## Phase 5 — Desktop Shell & CLI Polish (Weeks 17–21)

### Build Turns
- [ ] B5.1: TUI shell (three-panel: sidebar, canvas, right rail) — terminal fallback
- [ ] B5.2: Desktop app shell (Electron, local)
- [ ] B5.3: Six-pane canvas (terminal, code, doc, CAD, browser, notebook)
- [ ] B5.4: Agent roster + workflow queue in right rail
- [ ] B5.5: Mission dashboard (building / verifying / blocked status)
- [ ] B5.6: Health + error panels
- [ ] B5.7: Budget + token usage tracker
- [ ] B5.8: Approval/review UI for high-risk tool calls
- [ ] B5.9: Incident timeline view

### Verification Turns
- [ ] V5.1: TUI renders correctly in 80-col and 220-col terminals
- [ ] V5.2: Desktop shell state matches CLI state (same world graph)
- [ ] V5.3: Approval flow: agent requests high-risk tool → desktop shows prompt → engineer approves/denies → audit event correct
- [ ] V5.4: Budget tracker accurate within 1% of actual token usage
- [ ] V5.5: Rollback from incident timeline view restores correct state

### Run Gate
> Open `uclaw` in desktop mode, spawn two agents, run a mission, watch live in canvas, approve a high-risk action, verify mission completes and dashboard reflects final state.

---

## Phase 6 — Voice Layer (Weeks 22–24)

### Build Turns
- [ ] B6.1: Voice input capture (local Whisper or system STT)
- [ ] B6.2: Voice command parser → lead agent task dispatch
- [ ] B6.3: Voice feedback (TTS for agent status updates)
- [ ] B6.4: `uclaw voice` CLI mode

### Verification Turns
- [ ] V6.1: Voice command → correct mission action (10-command accuracy test)
- [ ] V6.2: Voice input never sent to external provider without explicit opt-in
- [ ] V6.3: Voice session audit-logged same as typed commands

### Run Gate
> Speak a mission brief aloud → lead agent briefs dev team → mission created and visible in desktop canvas.

---

## Phase 7 — Hardening & Production Readiness (Weeks 25–28)

### Build Turns
- [ ] B7.1: Full observability stack (metrics, structured logs, trace IDs)
- [ ] B7.2: Policy engine (YAML-defined rules, runtime enforcement)
- [ ] B7.3: Recovery paths (restore state, revert artifacts, tighten policies post-incident)
- [ ] B7.4: Multi-machine sync (optional, for team use)
- [ ] B7.5: Plugin/adapter SDK (for community integrations)
- [ ] B7.6: Documentation site (generated from vault + handbook)

### Verification Turns
- [ ] V7.1: Policy engine blocks all out-of-policy actions in fuzz test
- [ ] V7.2: Recovery path restores clean state after simulated incident
- [ ] V7.3: Multi-machine sync conflict resolution test
- [ ] V7.4: Full audit log coverage (every agent action has event)
- [ ] V7.5: Zero secrets in any log, config, or database field (automated scan)

### Run Gate
> Simulate an incident (agent takes out-of-policy action, mission corrupted), execute recovery path, confirm clean state, audit log shows full incident timeline. Zero secrets leaked.

---

## Milestone Summary

| Phase | Milestone | Gate |
|---|---|---|
| 0 | Foundation | `uclaw init` works |
| 1 | Agent Runtime | Agents spawn, tools gated, audit logs |
| 2 | Memory Vault | Graph reads/writes, Obsidian compat |
| 3 | Artifact System | Provenance scoring, sign-off chain |
| 4 | Verification Layer | Full mission flow with blocks + replay |
| 5 | Desktop Shell | Living control room visible |
| 6 | Voice Layer | Spoken mission dispatch |
| 7 | Hardened | Incident recovery, zero secrets leak |
