# Observability Architecture

## Layers

### 1. Structured Audit Log
- File: `~/.uclaw/audit.jsonl`
- Every agent action, tool call, approval, policy decision
- Fields: `event_id`, `trace_id`, `timestamp`, `agent_id`, `action`, `target`, `outcome`, `mission_id`
- Append-only; tamper detection via hash chaining (each event hashes prev)

### 2. Mission Dashboard
- Visible in desktop right rail and `uclaw status` CLI
- Shows per-mission: status (building / verifying / blocked / complete), active agents, pending checks

### 3. Health Panel
- Agent health: last heartbeat, current task, error count
- Provider health: latency, error rate, token usage per provider
- World state health: DB size, vault size, checkpoint count

### 4. Budget & Token Tracker
- Per-agent, per-mission, per-provider token counts
- Cost estimates based on provider pricing configs
- Current CLI supports reporting with `uclaw budget` and usage recording with `uclaw budget record --mission <id> --agent <id> --provider <name> --tokens <n> --cost <amount>`

### 5. Error Panel
- Aggregated errors from all agents in current session
- Error types: tool rejection, policy violation, provider timeout, verification failure
- One-click: show related audit events, navigate to blocked artifact

---

## Manual Controls

| Control | CLI | Desktop |
|---|---|---|
| Pause all agents | `uclaw pause --all` | Pause button in right rail |
| Pause one agent | `uclaw pause --agent <id>` | Agent context menu |
| Resume | `uclaw resume --agent <id>` | Resume button |
| Override block | `uclaw override --artifact <id>` | Override in review queue |
| Rollback mission | `uclaw mission rollback <id>` | Rollback in incident timeline |
| Tighten policy | `uclaw policy tighten --tool <name>` | Policy panel |

The current desktop is a local HTML/TUI renderer over shared runtime state, not yet a packaged desktop application with fully interactive controls.
