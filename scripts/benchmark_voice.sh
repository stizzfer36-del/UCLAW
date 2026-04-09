#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
tmp_home="$(mktemp -d)"
binary="$tmp_home/uclaw"
results_json="$repo_root/compliance/evidence/voice_benchmark_latest.json"
results_md="$repo_root/compliance/evidence/voice_benchmark_latest.md"
manifest="$repo_root/compliance/corpus/voice/manifest.json"
stt_command="bash $repo_root/scripts/offline-stt.sh {input}"

cleanup() {
  rm -rf "$tmp_home"
}
trap cleanup EXIT

mkdir -p "$repo_root/compliance/evidence"

(
  cd "$repo_root"
  UCLAW_HOME="$tmp_home/.uclaw" go build -o "$binary" ./cmd/uclaw
  UCLAW_HOME="$tmp_home/.uclaw" "$binary" init >/dev/null
)

python3 - "$binary" "$manifest" "$stt_command" "$results_json" "$results_md" "$tmp_home/.uclaw" <<'PY'
import json
import os
import subprocess
import sys
from datetime import date

binary, manifest_path, stt_command, results_json, results_md, home = sys.argv[1:]

with open(manifest_path, "r", encoding="utf-8") as fh:
    manifest = json.load(fh)

cases = []
correct = 0
for item in manifest["cases"]:
    cmd = [
        binary,
        "voice",
        "--audio",
        item["audio"],
        "--stt-command",
        stt_command,
    ]
    proc = subprocess.run(
        cmd,
        cwd=os.path.dirname(manifest_path) + "/../../..",
        env={**os.environ, "UCLAW_HOME": home},
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
        text=True,
        check=True,
    )
    result = json.loads(proc.stdout)
    transcript_ok = result.get("transcript", "").strip() == item["expected_transcript"]
    action_ok = result.get("action") == item["expected_action"]
    passed = transcript_ok and action_ok
    if passed:
        correct += 1
    cases.append(
        {
            "id": item["id"],
            "audio": item["audio"],
            "expected_transcript": item["expected_transcript"],
            "actual_transcript": result.get("transcript", "").strip(),
            "expected_action": item["expected_action"],
            "actual_action": result.get("action"),
            "passed": passed,
        }
    )

summary = {
    "generated_at": str(date.today()),
    "transcriber": "scripts/offline-stt.sh",
    "accuracy": correct / len(cases) if cases else 0.0,
    "cases_passed": correct,
    "cases_total": len(cases),
    "cases": cases,
}

with open(results_json, "w", encoding="utf-8") as fh:
    json.dump(summary, fh, indent=2)
    fh.write("\n")

with open(results_md, "w", encoding="utf-8") as fh:
    fh.write("# Voice Benchmark\n\n")
    fh.write(f"- Generated: {summary['generated_at']}\n")
    fh.write(f"- Deterministic offline transcriber: `scripts/offline-stt.sh`\n")
    fh.write(f"- Accuracy: {summary['cases_passed']}/{summary['cases_total']} ({summary['accuracy']:.0%})\n\n")
    fh.write("| Case | Expected Action | Actual Action | Expected Transcript | Actual Transcript | Pass |\n")
    fh.write("| --- | --- | --- | --- | --- | --- |\n")
    for case in cases:
        fh.write(
            f"| {case['id']} | {case['expected_action']} | {case['actual_action']} | "
            f"{case['expected_transcript']} | {case['actual_transcript']} | "
            f"{'yes' if case['passed'] else 'no'} |\n"
        )
PY
