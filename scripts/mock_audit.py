#!/usr/bin/env python3
import json
import pathlib
from datetime import date

REPO = pathlib.Path(__file__).resolve().parent.parent
MATRIX = REPO / "compliance" / "mappings" / "NIST_SSDF_MATRIX.md"
OUT_JSON = REPO / "compliance" / "evidence" / "mock_audit_latest.json"
OUT_MD = REPO / "compliance" / "evidence" / "mock_audit_latest.md"

rows = []
for line in MATRIX.read_text(encoding="utf-8").splitlines():
    if not line.startswith("|") or "Control" in line or "---" in line:
        continue
    parts = [part.strip() for part in line.strip("|").split("|")]
    if len(parts) < 3:
        continue
    control, requirement, refs = parts[:3]
    paths = [seg.strip().strip("`") for seg in refs.split(",")]
    checks = []
    for rel in paths:
        if not rel:
            continue
        if any(ch in rel for ch in "*?[]"):
            matches = list(REPO.glob(rel))
            if not matches and rel == "internal/*_test.go":
                matches = list((REPO / "internal").rglob("*_test.go"))
            present = any(match.exists() for match in matches)
        else:
            present = (REPO / rel).exists()
        checks.append({"path": rel, "present": present})
    rows.append(
        {
            "control": control,
            "requirement": requirement,
            "references": checks,
            "status": "pass" if checks and all(item["present"] for item in checks) else "fail",
        }
    )

summary = {
    "generated_at": str(date.today()),
    "matrix": str(MATRIX.relative_to(REPO)),
    "controls": rows,
    "passed": sum(1 for row in rows if row["status"] == "pass"),
    "failed": sum(1 for row in rows if row["status"] == "fail"),
}

OUT_JSON.parent.mkdir(parents=True, exist_ok=True)
with open(OUT_JSON, "w", encoding="utf-8") as fh:
    json.dump(summary, fh, indent=2)
    fh.write("\n")

with open(OUT_MD, "w", encoding="utf-8") as fh:
    fh.write("# Mock Audit\n\n")
    fh.write(f"- Generated: {summary['generated_at']}\n")
    fh.write(f"- Matrix: `{summary['matrix']}`\n")
    fh.write(f"- Passed controls: {summary['passed']}\n")
    fh.write(f"- Failed controls: {summary['failed']}\n\n")
    fh.write("| Control | Status | Evidence References |\n")
    fh.write("| --- | --- | --- |\n")
    for row in rows:
        refs = ", ".join(f"{item['path']}={'yes' if item['present'] else 'no'}" for item in row["references"])
        fh.write(f"| {row['control']} | {row['status']} | {refs} |\n")
