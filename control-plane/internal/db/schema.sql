CREATE TABLE IF NOT EXISTS agents (
    id         TEXT PRIMARY KEY,
    name       TEXT NOT NULL,
    status     TEXT NOT NULL DEFAULT 'offline',
    created_at DATETIME DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS scans (
    id            TEXT PRIMARY KEY,
    agent_id      TEXT NOT NULL REFERENCES agents(id),
    source_path   TEXT NOT NULL,
    status        TEXT NOT NULL DEFAULT 'pending',
    total_files   INTEGER DEFAULT 0,
    total_folders INTEGER DEFAULT 0,
    error         TEXT,
    created_at    DATETIME DEFAULT (datetime('now')),
    updated_at    DATETIME DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS scan_tree (
    id       INTEGER PRIMARY KEY AUTOINCREMENT,
    scan_id  TEXT NOT NULL REFERENCES scans(id),
    path     TEXT NOT NULL,
    is_dir   INTEGER NOT NULL,
    size     INTEGER DEFAULT 0,
    mod_time TEXT DEFAULT ''
);
