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

## ADR-006: Go for Core Runtime, JavaScript for Desktop Shell

**Decision:** Core (`world`, `agents`, `memory`, `artifacts`, `policies`) in Go. Desktop shell in JavaScript/Electron. CLI in Go.
**Reasoning:** Go: performance, single binary, strong stdlib for IPC/SQLite. Electron: broader packaging path in the current environment and no Rust toolchain dependency.
**Trade-off:** Two language contexts; mitigated by clean IPC boundary.
**Status:** Proposed

---

## ADR-007: SQLite Bridge Uses Local Python `sqlite3` in Development

**Decision:** The Go runtime invokes the host's local Python `sqlite3` module for SQLite access until a vendored Go SQLite driver is available locally.
**Reasoning:** This machine has Go and Python but no `sqlite3` CLI and no network access to fetch a Go driver. Using Python keeps state in a real local SQLite database and preserves the architecture's local-first guarantee.
**Trade-off:** The core runtime remains Go-controlled, but DB operations currently cross a local subprocess boundary.
**Status:** Accepted

---

## ADR-008: Desktop Shell Starts As Shared-State Local Web/TUI Renderers

**Decision:** The desktop shell is implemented first as a local HTML renderer plus terminal TUI backed by the same Go state loader used by the CLI, then packaged through Electron.
**Reasoning:** This preserves the shared-state architecture and keeps the packaged desktop path aligned with the available toolchain.
**Trade-off:** The current shell is still renderer-first, and local bundle verification depends on Electron dependencies being installed.
**Status:** Accepted

---

## ADR-009: Voice Phase Uses Local Transcript Dispatch In This Environment

**Decision:** Voice dispatch is implemented through local transcript input (`uclaw voice --text ...`) with the same mission creation and audit paths a speech-to-text frontend would call.
**Reasoning:** This machine can verify local voice command parsing, opt-in boundaries, and audit behavior without depending on unavailable microphone/STT system integration.
**Trade-off:** The current Phase 6 gate is satisfied through transcript-driven local dispatch rather than live microphone capture.
**Status:** Accepted
