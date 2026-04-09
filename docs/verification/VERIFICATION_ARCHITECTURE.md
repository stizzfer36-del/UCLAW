# Verification Architecture

## Principle

Verification is a first-class layer, not an afterthought.
Every significant artifact has associated checks before it can be signed off.
Planner teams create explicit verification branches as part of every mission.

---

## Verification Flow

```
Mission Start
    │
    ▼
 Planner (Lead Agent)
    │
    ├──▶ Dev Team (build artifacts)
    │         │
    │         ▼
    │    Artifact produced
    │         │
    └──▶ Verifier Team (independent agents)
              │
              ├── Verifier A: test execution
              ├── Verifier B: citation review
              └── Verifier C: policy compliance
                        │
                        ▼
                  All checks pass?
                  ┌────┴────┐
                 YES        NO
                  │          │
                  ▼          ▼
            Sign-off    Review Queue
            artifact    (blocked)
```

---

## Check Types

| Check | What it verifies | Run by |
|---|---|---|
| `test` | Code tests pass | Verifier Agent A |
| `citation_review` | Sources cited, completeness ≥80% | Verifier Agent B |
| `policy_compliance` | No policy violations in artifact | Verifier Agent C |
| `diff_review` | Code diff is coherent, no regressions | Lead Agent or human |
| `sign_off` | Human or lead agent final approval | Human / Lead |

---

## Test Matrix

Each mission defines a test matrix in its mission record:

```yaml
# missions/<id>/test_matrix.yaml
tests:
  - id: unit-auth
    type: unit
    target: core/auth/
    command: go test ./core/auth/...
    required: true

  - id: integration-api
    type: integration
    target: core/api/
    command: go test -tags integration ./core/api/...
    required: true

  - id: citation-check
    type: citation_review
    target: docs/
    min_completeness: 80
    required: true
```

---

## Replayable Workflows

- Every mission is checkpointed at each verification gate
- Stored in `~/.uclaw/checkpoints/<mission_id>/`
- Can re-run from any checkpoint: `uclaw mission replay <id> --from <checkpoint>`
- Enables: debugging failed verifications, auditing past runs, recovering from bad states

---

## Feedback Loop

Verification results feed back into agent handbooks and policies:

- If Verifier B consistently flags Agent X for poor citations → Agent X handbook updated with stricter citation rules
- If policy check fails repeatedly on a tool → tool risk level upgraded in `tools.yaml`
- Mission post-mortems auto-generate suggested handbook amendments
- Result: system gets stricter and more coherent over time, not looser
