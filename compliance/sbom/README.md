# SBOM

Generated SBOM artifacts should be stored here in SPDX or CycloneDX format.

Current repo state:

- `python3 scripts/generate_sbom.py` writes SPDX JSON documents for the Go and Node surfaces.
- CI regenerates SBOMs in `.github/workflows/security.yml`.
