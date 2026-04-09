CREATE TABLE IF NOT EXISTS artifacts (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    type TEXT NOT NULL,
    path TEXT NOT NULL,
    origin_agent TEXT NOT NULL,
    mission_id TEXT NOT NULL,
    created_at TEXT NOT NULL,
    verification_status TEXT NOT NULL CHECK(verification_status IN ('pending','in-review','verified','failed','disputed')),
    git_ref TEXT,
    trust_level TEXT NOT NULL CHECK(trust_level IN ('provenanced','partially-provenanced','unprovenanced')),
    claim_count INTEGER NOT NULL DEFAULT 1
);

CREATE TABLE IF NOT EXISTS artifact_sources (
    id TEXT PRIMARY KEY,
    artifact_id TEXT NOT NULL REFERENCES artifacts(id),
    source_type TEXT NOT NULL CHECK(source_type IN ('url','vault_node')),
    source_ref TEXT NOT NULL,
    cited_by TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS artifact_checks (
    id TEXT PRIMARY KEY,
    artifact_id TEXT NOT NULL REFERENCES artifacts(id),
    check_type TEXT NOT NULL,
    status TEXT NOT NULL CHECK(status IN ('pending','passed','failed')),
    run_by TEXT,
    run_at TEXT,
    details TEXT
);

CREATE TABLE IF NOT EXISTS artifact_signoffs (
    id TEXT PRIMARY KEY,
    artifact_id TEXT NOT NULL REFERENCES artifacts(id),
    signed_by TEXT NOT NULL,
    signed_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS artifact_versions (
    id TEXT PRIMARY KEY,
    artifact_id TEXT NOT NULL REFERENCES artifacts(id),
    git_ref TEXT NOT NULL,
    path TEXT NOT NULL,
    recorded_at TEXT NOT NULL
);
