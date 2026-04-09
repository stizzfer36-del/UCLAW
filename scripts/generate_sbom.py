#!/usr/bin/env python3
import json
import pathlib
import re
from datetime import date

REPO = pathlib.Path(__file__).resolve().parent.parent
OUT = REPO / "compliance" / "sbom"
OUT.mkdir(parents=True, exist_ok=True)


def load_json(path):
    with open(path, "r", encoding="utf-8") as fh:
        return json.load(fh)


def write_json(path, payload):
    with open(path, "w", encoding="utf-8") as fh:
        json.dump(payload, fh, indent=2)
        fh.write("\n")


def spdx_doc(name, packages):
    relationships = []
    root_id = packages[0]["SPDXID"]
    for package in packages[1:]:
        relationships.append(
            {
                "spdxElementId": root_id,
                "relationshipType": "DEPENDS_ON",
                "relatedSpdxElement": package["SPDXID"],
            }
        )
    return {
        "spdxVersion": "SPDX-2.3",
        "dataLicense": "CC0-1.0",
        "SPDXID": "SPDXRef-DOCUMENT",
        "name": name,
        "documentNamespace": f"https://uclaw.local/spdx/{name}/{date.today()}",
        "creationInfo": {
            "created": f"{date.today()}T00:00:00Z",
            "creators": ["Tool: scripts/generate_sbom.py"],
        },
        "packages": packages,
        "relationships": relationships,
    }


def package_entry(name, version, spdx_id, download="NOASSERTION"):
    return {
        "name": name,
        "SPDXID": spdx_id,
        "versionInfo": version or "NOASSERTION",
        "downloadLocation": download,
        "filesAnalyzed": False,
        "licenseConcluded": "NOASSERTION",
        "licenseDeclared": "NOASSERTION",
        "supplier": "NOASSERTION",
    }


def generate_go_sbom():
    go_mod = (REPO / "go.mod").read_text(encoding="utf-8")
    packages = [package_entry("uclaw", "0.1.0", "SPDXRef-Package-uclaw")]
    in_require_block = False
    for line in go_mod.splitlines():
        stripped = line.strip()
        if stripped.startswith("require ("):
            in_require_block = True
            continue
        if in_require_block and stripped == ")":
            in_require_block = False
            continue
        if not stripped or stripped.startswith("//"):
            continue
        match = None
        if in_require_block:
            match = re.match(r"([^\s]+)\s+([^\s]+)", stripped)
        elif stripped.startswith("require "):
            match = re.match(r"require\s+([^\s]+)\s+([^\s]+)", stripped)
        if not match:
            continue
        name, version = match.groups()
        packages.append(package_entry(name, version, f"SPDXRef-Package-{len(packages)}", f"https://pkg.go.dev/{name}"))
    write_json(OUT / "go.spdx.json", spdx_doc("uclaw-go", packages))


def generate_node_sbom():
    root = load_json(REPO / "package.json")
    packages = [package_entry(root["name"], root.get("version", "0.0.0"), "SPDXRef-Package-root")]
    desktop = load_json(REPO / "desktop" / "package.json")
    packages.append(package_entry(desktop["name"], desktop.get("version", "0.0.0"), "SPDXRef-Package-desktop"))
    lock_path = REPO / "package-lock.json"
    if lock_path.exists():
        lock = load_json(lock_path)
        for name, info in sorted(lock.get("packages", {}).items()):
            if not name or name == "":
                continue
            pkg_name = name.split("node_modules/")[-1]
            packages.append(package_entry(pkg_name, info.get("version", "0.0.0"), f"SPDXRef-Package-{len(packages)}"))
    write_json(OUT / "node.spdx.json", spdx_doc("uclaw-node", packages))


generate_go_sbom()
generate_node_sbom()
