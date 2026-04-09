# Security Architecture

## Core Principle

Agents are identities with scopes, not interchangeable prompts.
The habitat enforces trust boundaries at every layer.

---

## Trust Boundaries

| Boundary | Enforcement |
|---|---|
| User ↔ Agent | Agent cannot impersonate user; all agent actions attributed to agent ID |
| Agent ↔ Host OS | Path whitelists, no root by default, sandboxed tool calls |
| Agent ↔ External Providers | API keys in local env only; provider calls logged |
| Builder ↔ Verifier Teams | Separate agent identities; verifier cannot be same agent as builder |

---

## Secrets Management

- All secrets live in `.uclaw/.env` (gitignored)
- No secrets in `world.db`, configs, or agent handbooks
- Secrets accessed via `uclaw secrets get <key>` which logs the access
- Rotation is currently a manual operator process; there is no `uclaw secrets rotate` command yet

---

## Tool Security Model

```yaml
# core/policies/tools.yaml (example)
tools:
  - name: read_file
    risk_level: low
    requires_approval: false
    path_whitelist: ["$WORKSPACE"]

  - name: write_file
    risk_level: medium
    requires_approval: false
    path_whitelist: ["$WORKSPACE"]

  - name: delete_file
    risk_level: high
    requires_approval: true
    path_whitelist: ["$WORKSPACE"]

  - name: shell_exec
    risk_level: high
    requires_approval: true
    command_whitelist: ["go build", "go test", "npm run"]

  - name: deploy
    risk_level: critical
    requires_approval: true
    requires_human: true
```

---

## Audit Events

Every agent action emits a structured audit event to `~/.uclaw/audit.jsonl`:

```json
{
  "event_id": "evt-0001",
  "timestamp": "2026-04-09T06:01:00Z",
  "agent_id": "dev-agent-a",
  "action": "write_file",
  "target": "core/auth/auth.go",
  "outcome": "success",
  "mission_id": "mission-007",
  "tool": "write_file",
  "approval_required": false,
  "approval_granted": null
}
```

---

## Source Provenance as Security Property

- Artifacts must carry their sources (see Artifact System)
- Unprovenanced outputs are flagged lower trust — visible in UI and audit log
- Citation completeness checked by verifier team before sign-off
- Agents that consistently produce unprovenanced output have their trust score reduced
