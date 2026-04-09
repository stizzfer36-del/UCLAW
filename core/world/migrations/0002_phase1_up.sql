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
