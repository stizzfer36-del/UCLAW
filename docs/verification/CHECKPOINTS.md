# Checkpoints & Mission Replay

## What Is a Checkpoint

A checkpoint is a full snapshot of mission state at a verification gate:
- All artifact records at that point
- Agent states (active tasks, memory pointers)
- Verification check results so far
- World state graph snapshot
- Audit log up to that point

---

## Checkpoint Triggers

| Trigger | Description |
|---|---|
| Mission start | Baseline checkpoint |
| Dev team completes a phase | Mid-mission checkpoint |
| Verification gate reached | Pre-verification snapshot |
| Sign-off granted | Post-verification snapshot |
| Policy violation | Emergency checkpoint before block |
| Manual: `uclaw checkpoint save` | Engineer-triggered |

---

## Replay Commands

```bash
# List checkpoints for a mission
uclaw mission checkpoints <mission_id>

# Replay mission from a specific checkpoint
uclaw mission replay <mission_id> --from <checkpoint_id>

# Diff two checkpoints
uclaw mission diff <checkpoint_a> <checkpoint_b>

# Restore world state to a checkpoint (destructive)
uclaw world restore --checkpoint <checkpoint_id> --confirm
```

---

## Incident Timeline

For each mission, UCLAW maintains an incident timeline viewable in the desktop:

```
[06:00] Mission started by user
[06:01] Dev Agent A spawned, briefed on auth module
[06:05] Dev Agent A wrote core/auth/auth.go (artifact art-0001)
[06:06] Verifier A ran unit tests: PASSED
[06:07] Verifier B ran citation review: FAILED (completeness 62%)
[06:07] Mission BLOCKED — artifact art-0001 in review queue
[06:08] Dev Agent A notified, adding citations
[06:10] Verifier B re-ran citation review: PASSED (completeness 84%)
[06:11] Lead Agent signed off on art-0001
[06:11] Mission checkpoint saved
```
