# Architecture Decision Records

## ADR-001: Local-First State Graph

**Decision:** World state lives in local SQLite (`~/.uclaw/world.db`), not a cloud service.
**Reasoning:** Sovereignty, privacy, offline operation, zero vendor lock-in for state.
**Trade-off:** Multi-machine sync requires explicit setup (Phase 7).
**Status:** Accepted

---

## ADR-002: Agents as Identities, Not Prompts

**Decision:** Each agent has a unique ID, handbook, capability set, and audit emitter.
**Reasoning:** Security — you cannot enforce trust boundaries on interchangeable prompts.
**Trade-off:** Higher setup cost per agent vs. simple prompt injection.
**Status:** Accepted

---

## ADR-003: Separate Builder and Verifier Teams

**Decision:** An agent that builds an artifact cannot be the verifier of that artifact.
**Reasoning:** Prevents self-certification; mirrors human engineering review practices.
**Trade-off:** Requires at least two agent identities per mission.
**Status:** Accepted

---

## ADR-004: Markdown-Native Vault

**Decision:** Memory vault uses markdown + frontmatter as the storage format.
**Reasoning:** Human-readable, Obsidian-compatible, durable, no proprietary format lock-in.
**Trade-off:** Slightly slower than pure DB for large graph traversals; mitigated by SQLite index.
**Status:** Accepted

---

## ADR-005: Verification Feedback into Handbooks

**Decision:** Failed verification checks automatically generate handbook amendment suggestions.
**Reasoning:** System should get stricter over time, not drift toward laxity.
**Trade-off:** Handbook amendments need engineer review before applying (not auto-applied).
**Status:** Accepted

---

## ADR-006: Go for Core Runtime, TypeScript for Desktop Shell

**Decision:** Core (`world`, `agents`, `memory`, `artifacts`, `policies`) in Go. Desktop shell in TypeScript (Tauri). CLI in Go (cobra).
**Reasoning:** Go: performance, single binary, strong stdlib for IPC/SQLite. TypeScript/Tauri: rich UI, web component ecosystem, local-only.
**Trade-off:** Two language contexts; mitigated by clean IPC boundary.
**Status:** Proposed
