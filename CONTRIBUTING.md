# Contributing to UCLAW

## Philosophy

Every contribution to UCLAW must itself follow UCLAW principles:
- **Provenanced:** changes cite their reasoning (in PR description or decision doc)
- **Verified:** every PR must pass its test matrix before merge
- **Audited:** significant design changes require an ADR in `docs/roadmap/DECISIONS.md`

## Workflow

1. Open an issue describing the change and its motivation
2. Create a branch: `feature/<phase>-<description>` or `fix/<description>`
3. Build the change
4. Write or update tests in `verification/`
5. Ensure all phase verification turns pass
6. Open a PR with:
   - Phase reference (e.g. "Phase 1, B1.3")
   - Test matrix results
   - Any new ADRs
7. A separate reviewer must approve (builder ≠ verifier, per ADR-003)

## Branch Naming

```
feature/phase0-world-schema
feature/phase1-agent-runtime
fix/phase2-vault-edge-traversal
docs/adr-007-voice-layer
```

## Test Commands

```bash
# Run all Go tests
go test ./...

# Run verification suite for a phase
bash verification/run_phase.sh <phase_number>

# Lint
golangci-lint run
```
