# Mock Audit

- Generated: 2026-04-09
- Matrix: `compliance/mappings/NIST_SSDF_MATRIX.md`
- Passed controls: 9
- Failed controls: 0

| Control | Status | Evidence References |
| --- | --- | --- |
| PO.1 Define security requirements | pass | compliance/policies/=yes |
| PO.3 Define roles and responsibilities | pass | compliance/policies/=yes |
| PO.5 Protect development environments | pass | docs/security/SECURITY_ARCHITECTURE.md=yes |
| PS.2 Review software for vulnerabilities | pass | compliance/findings/=yes, compliance/evidence/=yes |
| PW.4 Manage third-party software | pass | go.mod=yes, package.json=yes, compliance/sbom/=yes |
| PW.5 Follow secure coding practices | pass | compliance/policies/=yes, internal/*_test.go=yes |
| PW.8 Reuse trusted software | pass | core/policies/=yes, desktop/=yes |
| RV.1 Identify and confirm vulnerabilities | pass | compliance/findings/=yes, .github/workflows/security.yml=yes |
| RV.3 Analyze and remediate vulnerabilities | pass | compliance/findings/=yes, compliance/exceptions/=yes |
