-- UCLAW World State Schema Snapshot
-- The authoritative schema is the migration set under core/world/migrations/.
-- This file mirrors the currently implemented Phase 0-4 schema.

CREATE TABLE IF NOT EXISTS world (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    created_at TEXT NOT NULL,
    vault_path TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS offices (
    id TEXT PRIMARY KEY,
    world_id TEXT NOT NULL REFERENCES world(id),
    name TEXT NOT NULL,
    description TEXT,
    created_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS teams (
    id TEXT PRIMARY KEY,
    office_id TEXT NOT NULL REFERENCES offices(id),
    name TEXT NOT NULL,
    role TEXT NOT NULL CHECK(role IN ('dev','verify','research','ops','lead')),
    lead_agent_id TEXT,
    created_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS members (
    id TEXT PRIMARY KEY,
    team_id TEXT NOT NULL REFERENCES teams(id),
    name TEXT NOT NULL,
    type TEXT NOT NULL CHECK(type IN ('human','agent')),
    handbook_path TEXT,
    created_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS capabilities (
    id TEXT PRIMARY KEY,
    member_id TEXT NOT NULL REFERENCES members(id),
    tool_name TEXT NOT NULL,
    granted_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS path_whitelists (
    id TEXT PRIMARY KEY,
    member_id TEXT NOT NULL REFERENCES members(id),
    path TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS machines (
    id TEXT PRIMARY KEY,
    member_id TEXT NOT NULL REFERENCES members(id),
    hostname TEXT NOT NULL,
    os TEXT
);

CREATE TABLE IF NOT EXISTS rooms (
    id TEXT PRIMARY KEY,
    machine_id TEXT NOT NULL REFERENCES machines(id),
    name TEXT NOT NULL,
    type TEXT NOT NULL CHECK(type IN ('mission','workspace','archive'))
);

CREATE TABLE IF NOT EXISTS missions (
    id TEXT PRIMARY KEY,
    room_id TEXT NOT NULL REFERENCES rooms(id),
    title TEXT NOT NULL,
    status TEXT NOT NULL CHECK(status IN ('active','blocked','complete','failed','archived')),
    created_by TEXT NOT NULL,
    assigned_to TEXT,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS agent_profiles (
    id TEXT PRIMARY KEY,
    member_id TEXT NOT NULL REFERENCES members(id),
    role TEXT NOT NULL,
    provider TEXT NOT NULL,
    status TEXT NOT NULL,
    trust_score INTEGER NOT NULL DEFAULT 100,
    created_at TEXT NOT NULL,
    retired_at TEXT
);

CREATE TABLE IF NOT EXISTS approval_requests (
    id TEXT PRIMARY KEY,
    agent_id TEXT NOT NULL,
    tool_name TEXT NOT NULL,
    target TEXT,
    status TEXT NOT NULL CHECK(status IN ('pending','approved','denied')),
    requested_at TEXT NOT NULL,
    decided_at TEXT,
    decided_by TEXT
);

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

CREATE TABLE IF NOT EXISTS review_queue (
    id TEXT PRIMARY KEY,
    artifact_id TEXT NOT NULL,
    reason TEXT NOT NULL,
    status TEXT NOT NULL CHECK(status IN ('open','overridden','resolved')),
    created_at TEXT NOT NULL,
    resolved_at TEXT
);

CREATE TABLE IF NOT EXISTS checkpoints (
    id TEXT PRIMARY KEY,
    mission_id TEXT NOT NULL REFERENCES missions(id),
    trigger TEXT NOT NULL,
    created_at TEXT NOT NULL,
    snapshot_path TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS timeline_events (
    id TEXT PRIMARY KEY,
    mission_id TEXT NOT NULL REFERENCES missions(id),
    event_type TEXT NOT NULL,
    message TEXT NOT NULL,
    created_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS handbook_amendments (
    id TEXT PRIMARY KEY,
    agent_id TEXT NOT NULL,
    reason TEXT NOT NULL,
    path TEXT NOT NULL,
    created_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS token_usage (
    id TEXT PRIMARY KEY,
    mission_id TEXT,
    agent_id TEXT,
    provider TEXT NOT NULL,
    tokens INTEGER NOT NULL,
    cost REAL NOT NULL,
    created_at TEXT NOT NULL
);
