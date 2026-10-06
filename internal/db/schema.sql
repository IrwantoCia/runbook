CREATE TABLE IF NOT EXISTS users (
 id TEXT PRIMARY KEY, email TEXT UNIQUE NOT NULL, password_hash TEXT NOT NULL, created_at TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS sessions (
 id TEXT PRIMARY KEY, user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
 token_hash TEXT UNIQUE NOT NULL, expires_at TEXT NOT NULL, created_at TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS api_keys (
 id TEXT PRIMARY KEY, name TEXT NOT NULL, key_hash TEXT UNIQUE NOT NULL, prefix TEXT NOT NULL,
 created_at TEXT NOT NULL, last_used_at TEXT, revoked_at TEXT
);
CREATE TABLE IF NOT EXISTS folders (
 id TEXT PRIMARY KEY, name TEXT NOT NULL, parent_id TEXT REFERENCES folders(id) ON DELETE RESTRICT,
 created_at TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS runbooks (
 id TEXT PRIMARY KEY, slug TEXT UNIQUE NOT NULL, title TEXT NOT NULL, body TEXT NOT NULL,
 status TEXT NOT NULL DEFAULT 'draft' CHECK(status IN ('draft','published')),
 created_by TEXT NOT NULL REFERENCES users(id), created_at TEXT NOT NULL, updated_at TEXT NOT NULL, published_at TEXT,
 folder_id TEXT REFERENCES folders(id) ON DELETE RESTRICT
);
