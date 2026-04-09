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
