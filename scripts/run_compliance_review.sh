#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
evidence_dir="$repo_root/compliance/evidence"
findings_dir="$repo_root/compliance/findings"
today="$(date +%F)"

mkdir -p "$evidence_dir" "$findings_dir"

(
  cd "$repo_root"
  go test ./... >"$evidence_dir/go_test_latest.log"
  go vet ./... >"$evidence_dir/go_vet_latest.log" 2>&1
  node desktop/build.js --doctor >"$evidence_dir/desktop_doctor_latest.json"
  if rg -n "secret-anthropic|secret-openrouter" . \
    --glob '!compliance/evidence/**' \
    --glob '!.git/**' \
    --glob '!internal/testingx/**' \
    --glob '!**/*_test.go' \
    --glob '!internal/hardening/hardening.go' \
    --glob '!scripts/run_compliance_review.sh' >"$evidence_dir/secret_scan_latest.log"; then
    echo "secret-like values found" >&2
    exit 1
  fi
  python3 scripts/generate_sbom.py
  python3 scripts/mock_audit.py
  bash scripts/benchmark_voice.sh
)

cat >"$findings_dir/findings.json" <<JSON
{
  "generated_at": "$today",
  "findings": []
}
JSON

cat >"$findings_dir/findings.md" <<MD
# Findings Register

Generated from local review evidence on $today.

No open normalized findings are currently tracked from:

- \`compliance/evidence/go_test_latest.log\`
- \`compliance/evidence/go_vet_latest.log\`
- \`compliance/evidence/secret_scan_latest.log\`
- \`compliance/evidence/voice_benchmark_latest.json\`
- \`compliance/evidence/mock_audit_latest.json\`

Evidence and review outputs are machine-regenerable via \`bash scripts/run_compliance_review.sh\`.
MD
