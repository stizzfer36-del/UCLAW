-- UCLAW World State Schema — SPEC-1 Phase 1
-- Implements full spec data model: missions, work_packets, agent_sessions,
-- terminals, artifacts, verification_runs, checkpoints, audit_events

-- ── bootstrap tables (kept for EnsureWorld hierarchy) ───────────────────────
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

-- ── core spec tables ─────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS missions (
    id TEXT PRIMARY KEY,
    parent_id TEXT,
    room_id TEXT NOT NULL REFERENCES rooms(id),
    title TEXT NOT NULL,
    goal TEXT NOT NULL DEFAULT '',
    state TEXT NOT NULL DEFAULT 'active'
        CHECK(state IN ('active','blocked','complete','failed','archived','rolled_back')),
    workspace_id TEXT NOT NULL DEFAULT 'ws-default',
    permissions_json TEXT NOT NULL DEFAULT '{}',
    budget_json TEXT NOT NULL DEFAULT '{"tokens":0,"cost":0}',
    routing_json TEXT NOT NULL DEFAULT '{"model":"default"}',
    created_by TEXT NOT NULL,
    assigned_to TEXT,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS work_packets (
    id TEXT PRIMARY KEY,
    mission_id TEXT NOT NULL REFERENCES missions(id),
    agent_id TEXT NOT NULL DEFAULT 'unassigned',
    kind TEXT NOT NULL DEFAULT 'task',
    input_json TEXT NOT NULL DEFAULT '{}',
    state TEXT NOT NULL DEFAULT 'queued'
        CHECK(state IN ('queued','planned','running','waiting_tool','waiting_human',
                        'blocked','verified','failed','rolled_back','completed')),
    priority INTEGER NOT NULL DEFAULT 0,
    idempotency_key TEXT NOT NULL DEFAULT '',
    started_at TEXT,
    finished_at TEXT
);

CREATE TABLE IF NOT EXISTS agent_sessions (
    id TEXT PRIMARY KEY,
    mission_id TEXT NOT NULL REFERENCES missions(id),
    agent_type TEXT NOT NULL DEFAULT 'primary',
    pack_id TEXT,
    model_profile TEXT NOT NULL DEFAULT 'default',
    context_window_state_json TEXT NOT NULL DEFAULT '{}',
    state TEXT NOT NULL DEFAULT 'idle'
        CHECK(state IN ('idle','running','waiting','suspended','terminated')),
    worktree_path TEXT,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS terminals (
    id TEXT PRIMARY KEY,
    mission_id TEXT REFERENCES missions(id),
    worktree_path TEXT,
    pty_ref TEXT NOT NULL DEFAULT '',
    layout_slot TEXT NOT NULL DEFAULT '1x1',
    state TEXT NOT NULL DEFAULT 'active'
        CHECK(state IN ('active','suspended','closed')),
    scrollback_path TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS artifacts (
    id TEXT PRIMARY KEY,
    mission_id TEXT NOT NULL,
    class TEXT NOT NULL DEFAULT 'source_code'
        CHECK(class IN ('source_code','document','spec','dataset','notebook',
                        'design_asset','cad_metadata','release_bundle',
                        'research_report','workflow_output')),
    title TEXT NOT NULL DEFAULT '',
    uri TEXT NOT NULL DEFAULT '',
    hash TEXT NOT NULL DEFAULT '',
    schema_version TEXT NOT NULL DEFAULT '1',
    status TEXT NOT NULL DEFAULT 'pending'
        CHECK(status IN ('pending','in-review','verified','failed','disputed')),
    trust_level TEXT NOT NULL DEFAULT 'unprovenanced'
        CHECK(trust_level IN ('provenanced','partially-provenanced','unprovenanced')),
    metadata_json TEXT NOT NULL DEFAULT '{}',
    origin_agent TEXT NOT NULL DEFAULT 'cli',
    git_ref TEXT,
    claim_count INTEGER NOT NULL DEFAULT 1,
    created_at TEXT NOT NULL
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

CREATE TABLE IF NOT EXISTS verification_runs (
    id TEXT PRIMARY KEY,
    artifact_id TEXT NOT NULL REFERENCES artifacts(id),
    verifier_type TEXT NOT NULL,
    status TEXT NOT NULL CHECK(status IN ('pending','running','passed','failed','skipped')),
    score REAL,
    output_json TEXT NOT NULL DEFAULT '{}',
    started_at TEXT NOT NULL,
    finished_at TEXT
);

CREATE TABLE IF NOT EXISTS checkpoints (
    id TEXT PRIMARY KEY,
    mission_id TEXT NOT NULL REFERENCES missions(id),
    reason TEXT NOT NULL,
    journal_offset INTEGER NOT NULL DEFAULT 0,
    snapshot_path TEXT NOT NULL,
    created_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS audit_events (
    seq INTEGER PRIMARY KEY AUTOINCREMENT,
    trace_id TEXT NOT NULL DEFAULT '',
    mission_id TEXT,
    actor_type TEXT NOT NULL DEFAULT 'system',
    actor_id TEXT NOT NULL DEFAULT 'system',
    event_type TEXT NOT NULL,
    payload_json TEXT NOT NULL DEFAULT '{}',
    created_at TEXT NOT NULL
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

CREATE TABLE IF NOT EXISTS review_queue (
    id TEXT PRIMARY KEY,
    artifact_id TEXT NOT NULL,
    reason TEXT NOT NULL,
    status TEXT NOT NULL CHECK(status IN ('open','overridden','resolved')),
    created_at TEXT NOT NULL,
    resolved_at TEXT
);

CREATE TABLE IF NOT EXISTS timeline_events (
    id TEXT PRIMARY KEY,
    mission_id TEXT NOT NULL,
    event_type TEXT NOT NULL,
    message TEXT NOT NULL,
    created_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS agent_profiles (
    id TEXT PRIMARY KEY,
    member_id TEXT NOT NULL,
    role TEXT NOT NULL,
    provider TEXT NOT NULL,
    status TEXT NOT NULL,
    trust_score INTEGER NOT NULL DEFAULT 100,
    created_at TEXT NOT NULL,
    retired_at TEXT
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

-- ── mission journal (no-turns-missed rule) ────────────────────────────────────
CREATE TABLE IF NOT EXISTS mission_journal (
    offset INTEGER PRIMARY KEY AUTOINCREMENT,
    mission_id TEXT NOT NULL,
    packet_id TEXT,
    event_type TEXT NOT NULL,
    payload_json TEXT NOT NULL DEFAULT '{}',
    idempotency_key TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL
);

-- ── indexes ───────────────────────────────────────────────────────────────────
CREATE INDEX IF NOT EXISTS idx_missions_state ON missions(state);
CREATE INDEX IF NOT EXISTS idx_work_packets_mission ON work_packets(mission_id);
CREATE INDEX IF NOT EXISTS idx_work_packets_state ON work_packets(state);
CREATE INDEX IF NOT EXISTS idx_agent_sessions_mission ON agent_sessions(mission_id);
CREATE INDEX IF NOT EXISTS idx_artifacts_mission ON artifacts(mission_id);
CREATE INDEX IF NOT EXISTS idx_artifacts_status ON artifacts(status);
CREATE INDEX IF NOT EXISTS idx_verification_runs_artifact ON verification_runs(artifact_id);
CREATE INDEX IF NOT EXISTS idx_audit_events_mission ON audit_events(mission_id);
CREATE INDEX IF NOT EXISTS idx_audit_events_trace ON audit_events(trace_id);
CREATE INDEX IF NOT EXISTS idx_journal_mission ON mission_journal(mission_id);
CREATE INDEX IF NOT EXISTS idx_timeline_mission ON timeline_events(mission_id);
