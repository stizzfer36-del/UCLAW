# Secure Development Policy

## Scope

In scope:

- repository source under `cmd/`, `internal/`, `desktop/`, `core/`, `docs/`, `scripts/`
- default branch and any branch intended to merge into release or production use
- local build flow, test flow, generated desktop assets, and sync artifacts
- local state stored under `~/.uclaw/`

Out of scope until implemented:

- hosted control planes
- cloud deployment environments
- managed identity providers

## Baselines

- Primary: NIST SP 800-218 Secure Software Development Framework
- Secondary mapping: OWASP ASVS, OWASP Top 10, FedRAMP-style evidence expectations

## Standards

- All code changes require peer review for security-sensitive paths.
- New runtime actions must emit structured audit events.
- Secrets must never be committed, logged, or stored in synced state bundles.
- External provider use must be explicit and auditable.
- Dependency additions must be justified and reviewed for supply-chain risk.
- Security regressions block release unless formally accepted in `compliance/exceptions/`.

## Evidence

Required evidence for high-risk changes:

- tests and verification output
- audit log proof where behavior changed
- threat-model updates for new trust boundaries
- findings updates and remediation links
- exceptions when risk is accepted instead of remediated
