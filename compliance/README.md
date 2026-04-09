# Compliance Layout

This directory is the evidence spine for UCLAW's secure-development and audit process.

- `policies/`: secure development policy and standards
- `findings/`: machine-readable and human-readable findings register
- `exceptions/`: approved risk acceptances
- `sbom/`: generated software bills of materials and supplier notes
- `evidence/`: collected scanner output, test logs, timing runs, and audit artifacts
- `training/`: secure coding and review records
- `mappings/`: control-to-evidence traceability

Regenerate current machine-produced artifacts with:

- `bash scripts/run_compliance_review.sh`
- `python3 scripts/generate_sbom.py`
- `python3 scripts/mock_audit.py`
- `bash scripts/benchmark_voice.sh`
