# Agent Runtime Architecture

## Agent as First-Class Identity

Every agent in UCLAW is not an interchangeable prompt — it is a worker with:

- **Identity:** unique ID, name, team membership, role
- **Engineering Handbook:** markdown file encoding citation rules, reasoning style, log format, escalation rules
- **Capability Set:** declared list of tools this agent may call
- **Path Whitelist:** filesystem paths this agent may read/write
- **Provider Binding:** which LLM backend this agent uses
- **Audit Emitter:** every action produces a structured audit event
- **Private Terminal:** dedicated shell session, isolated from other agents

---

## Team Orchestration Model

```
Lead Agent (Planner)
├── receives mission brief from user
├── decomposes into tasks
├── spawns Dev Team sub-agents
│   ├── Dev Agent A: implementation
│   ├── Dev Agent B: research + sourcing
│   └── Dev Agent C: documentation
└── spawns Verifier Team sub-agents
    ├── Verifier A: test execution
    ├── Verifier B: citation review
    └── Verifier C: policy compliance check
```

---

## Agent Lifecycle

1. **Spawn** — Lead or user creates agent, assigns handbook + capabilities
2. **Brief** — Agent receives mission context from memory vault
3. **Execute** — Agent runs tasks within its capability scope
4. **Checkpoint** — Agent writes progress to mission log + memory vault
5. **Handoff** — Agent signals completion, passes artifacts to verifier or lead
6. **Audit** — All actions already logged to structured audit stream
7. **Retire** — Agent session closed, state preserved in vault

---

## Engineering Handbook Schema

Each agent's handbook is a markdown file at `~/.uclaw/agents/<id>/handbook.md`:

```markdown
# Agent Handbook: <name>

## Role
<description>

## Citation Rules
- Always cite source URLs inline
- Minimum citation completeness: 80%
- Flag unverifiable claims with [UNVERIFIED]

## Reasoning Style
- Think step by step before acting
- Prefer reversible actions
- Escalate if confidence < 60%

## Log Format
- Action: <verb> <object>
- Source: <url or path>
- Confidence: <0-100>
- Timestamp: <ISO8601>

## Escalation Rules
- Tool call requires path outside whitelist → request approval
- Risky action (delete, deploy, send) → human review
- Conflicting instructions → flag to lead agent
```

---

## Tool Call Security

- Tools are registered in `core/policies/tools.yaml`
- Each tool declares: `name`, `risk_level` (low/medium/high/critical), `requires_approval`
- High/critical tools require explicit approval in the desktop UI or `uclaw approve <action_id>`
- Unapproved tool calls are rejected and logged as audit events
