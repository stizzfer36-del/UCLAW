# NIST SSDF Control Mapping

| SSDF Practice | Repo Control | Evidence Location |
|---|---|---|
| PO.1 Define security requirements | `compliance/policies/SECURE_DEVELOPMENT_POLICY.md` | `compliance/policies/` |
| PO.3 Define roles and responsibilities | repo owners and reviewer expectations in policy | `compliance/policies/` |
| PO.5 Protect development environments | local-first state, secrets isolation, policy gates | `docs/security/SECURITY_ARCHITECTURE.md` |
| PS.2 Review software for vulnerabilities | test suite, audit checks, findings register | `compliance/findings/`, `compliance/evidence/` |
| PW.4 Manage third-party software | package manifests plus SBOM location | `go.mod`, `package.json`, `compliance/sbom/` |
| PW.5 Follow secure coding practices | secure development policy, code review, tests | `compliance/policies/`, `internal/*_test.go` |
| PW.8 Reuse trusted software | controlled policy registry and local runtime assets | `core/policies/`, `desktop/` |
| RV.1 Identify and confirm vulnerabilities | findings register and security workflow | `compliance/findings/`, `.github/workflows/security.yml` |
| RV.3 Analyze and remediate vulnerabilities | remediation branches and evidence updates | `compliance/findings/`, `compliance/exceptions/` |
