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
