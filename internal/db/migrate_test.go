package db

import (
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"
)

func TestMigrateFreshSchema(t *testing.T) {
	database := openTestDatabase(t)
	if err := Migrate(database); err != nil {
		t.Fatal(err)
	}
	if err := Migrate(database); err != nil {
		t.Fatalf("repeat migration: %v", err)
	}
	var folderColumn int
	if err := database.QueryRow(`SELECT COUNT(*) FROM pragma_table_info('runbooks') WHERE name='folder_id'`).Scan(&folderColumn); err != nil {
		t.Fatal(err)
	}
	if folderColumn != 1 {
		t.Fatalf("folder_id column count = %d, want 1", folderColumn)
	}
}

func TestMigrateLegacySchemaPreservesDataAndIsRepeatable(t *testing.T) {
	database := openTestDatabase(t)
	_, err := database.Exec(`
CREATE TABLE users (
 id TEXT PRIMARY KEY, email TEXT UNIQUE NOT NULL, password_hash TEXT NOT NULL, created_at TEXT NOT NULL
);
CREATE TABLE sessions (
 id TEXT PRIMARY KEY, user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
 token_hash TEXT UNIQUE NOT NULL, expires_at TEXT NOT NULL, created_at TEXT NOT NULL
);
CREATE TABLE api_keys (
 id TEXT PRIMARY KEY, name TEXT NOT NULL, key_hash TEXT UNIQUE NOT NULL, prefix TEXT NOT NULL,
 created_at TEXT NOT NULL, last_used_at TEXT, revoked_at TEXT
);
CREATE TABLE runbooks (
 id TEXT PRIMARY KEY, slug TEXT UNIQUE NOT NULL, title TEXT NOT NULL, body TEXT NOT NULL,
 status TEXT NOT NULL DEFAULT 'draft' CHECK(status IN ('draft','review','published')),
 created_by TEXT NOT NULL REFERENCES users(id), created_at TEXT NOT NULL, updated_at TEXT NOT NULL, published_at TEXT
);
INSERT INTO users VALUES ('u1','owner@example.com','hash','2020-01-01T00:00:00Z');
INSERT INTO sessions VALUES ('s1','u1','session-hash','2030-01-01T00:00:00Z','2020-01-01T00:00:00Z');
INSERT INTO api_keys VALUES ('k1','legacy key','key-hash','prefix','2020-01-01T00:00:00Z',NULL,NULL);
INSERT INTO runbooks VALUES ('draft-id','draft-slug','Draft title','draft body','draft','u1','2020-01-01T00:00:00Z','2020-01-02T00:00:00Z',NULL);
INSERT INTO runbooks VALUES ('review-id','review-slug','Review title','review body','review','u1','2020-02-01T00:00:00Z','2020-02-02T00:00:00Z',NULL);
INSERT INTO runbooks VALUES ('published-id','published-slug','Published title','published body','published','u1','2020-03-01T00:00:00Z','2020-03-02T00:00:00Z','2020-03-03T00:00:00Z');`)
	if err != nil {
		t.Fatal(err)
	}
	if err := Migrate(database); err != nil {
		t.Fatalf("migrate legacy schema: %v", err)
	}
	if err := Migrate(database); err != nil {
		t.Fatalf("repeat migration: %v", err)
	}

	rows, err := database.Query(`SELECT id,slug,title,body,status,folder_id FROM runbooks ORDER BY id`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	want := []struct{ id, slug, title, body, status string }{
		{"draft-id", "draft-slug", "Draft title", "draft body", "draft"},
		{"published-id", "published-slug", "Published title", "published body", "published"},
		{"review-id", "review-slug", "Review title", "review body", "draft"},
	}
	for _, expected := range want {
		var id, slug, title, body, status string
		var folderID sql.NullString
		if !rows.Next() {
			t.Fatalf("missing migrated runbook %q: %v", expected.id, rows.Err())
		}
		if err := rows.Scan(&id, &slug, &title, &body, &status, &folderID); err != nil {
			t.Fatal(err)
		}
		if id != expected.id || slug != expected.slug || title != expected.title || body != expected.body || status != expected.status || folderID.Valid {
			t.Errorf("migrated row = (%q,%q,%q,%q,%q,folder valid=%v), want (%q,%q,%q,%q,%q,root)", id, slug, title, body, status, folderID.Valid, expected.id, expected.slug, expected.title, expected.body, expected.status)
		}
	}
	if rows.Next() {
		t.Fatal("unexpected extra migrated runbook")
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	for _, table := range []string{"sessions", "api_keys"} {
		var count int
		if err := database.QueryRow(`SELECT COUNT(*) FROM ` + table).Scan(&count); err != nil {
			t.Fatal(err)
		}
		if count != 1 {
			t.Errorf("%s rows = %d, want 1", table, count)
		}
	}
	var violations int
	if err := database.QueryRow(`SELECT COUNT(*) FROM pragma_foreign_key_check`).Scan(&violations); err != nil {
		t.Fatal(err)
	}
	if violations != 0 {
		t.Fatalf("foreign key violations = %d", violations)
	}
}

func openTestDatabase(t *testing.T) *sql.DB {
	t.Helper()
	database, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	database.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = database.Close() })
	if _, err := database.Exec(`PRAGMA foreign_keys=ON`); err != nil {
		t.Fatal(err)
	}
	return database
}
