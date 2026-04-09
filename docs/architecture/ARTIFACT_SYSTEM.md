# Artifact System Architecture

## What Is an Artifact

An artifact is any significant output produced by an agent or human in UCLAW:
code files, documents, slides, CAD specs, test reports, screenshots, prompt files.

Every artifact carries:
- **Origin:** which agent or human created it
- **Mission:** which mission it belongs to
- **Sources:** citations (URLs, vault nodes) used to produce it
- **Verification Status:** pending | in-review | verified | failed | disputed
- **Sign-Off Chain:** list of agents/humans who approved it
- **Git Reference:** commit SHA or branch if code-backed
- **Trust Level:** provenanced | partially-provenanced | unprovenanced

---

## Artifact Record Schema

```json
{
  "id": "art-0001",
  "title": "Auth module implementation",
  "type": "code",
  "path": "core/auth/auth.go",
  "origin_agent": "dev-agent-a",
  "mission_id": "mission-007",
  "created_at": "2026-04-09T06:00:00Z",
  "sources": [
    {"url": "https://pkg.go.dev/crypto", "cited_by": "dev-agent-a"},
    {"vault_node": "decisions/auth-design.md", "cited_by": "dev-agent-a"}
  ],
  "verification_status": "in-review",
  "sign_off_chain": [],
  "git_ref": "feature/auth-module",
  "trust_level": "provenanced",
  "checks": [
    {"type": "test", "status": "passed", "run_by": "verifier-a"},
    {"type": "citation_review", "status": "pending", "run_by": "verifier-b"},
    {"type": "policy_compliance", "status": "passed", "run_by": "verifier-c"}
  ]
}
```

---

## Trust Level Rules

| Level | Condition | UI Display |
|---|---|---|
| `provenanced` | ≥80% of claims cite verifiable sources | Green badge |
| `partially-provenanced` | 40–79% cited | Yellow badge |
| `unprovenanced` | <40% cited or no sources | Red badge + warning |

---

## Artifact CLI

```bash
# List artifacts for current mission
uclaw artifact list

# Show artifact details + checks
uclaw artifact show <id>

# Manually flag an artifact for re-verification
uclaw artifact flag <id> --reason "source changed"

# Sign off on a verified artifact
uclaw artifact sign <id>

# Revert artifact to previous version
uclaw artifact revert <id> --to <git-sha>
```
