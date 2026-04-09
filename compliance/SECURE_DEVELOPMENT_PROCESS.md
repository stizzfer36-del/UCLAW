# Secure Development Process

1. Define or update the threat model for new trust boundaries.
2. Implement changes with audit coverage and tests.
3. Run local verification: `go test ./...`, desktop asset build, and secret scan.
4. Record any discovered issues in `compliance/findings/`.
5. Require peer review for security-sensitive changes before merge.
6. Re-run verification after merge and archive evidence under `compliance/evidence/`.
7. Use `compliance/exceptions/` for residual risk acceptance only after compensating controls are documented.
