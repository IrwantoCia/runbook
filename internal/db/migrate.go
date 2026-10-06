package db

import (
	"database/sql"
	_ "embed"
	"strings"
)

//go:embed schema.sql
var schema string

func Migrate(database *sql.DB) error {
	if _, err := database.Exec(schema); err != nil {
		return err
	}
	var definition string
	if err := database.QueryRow(`SELECT sql FROM sqlite_master WHERE type='table' AND name='runbooks'`).Scan(&definition); err != nil {
		return err
	}
	if !strings.Contains(definition, "'review'") {
		_, err := database.Exec(`ALTER TABLE runbooks ADD COLUMN folder_id TEXT REFERENCES folders(id) ON DELETE RESTRICT`)
		if err != nil && !strings.Contains(err.Error(), "duplicate column") {
			return err
		}
		return nil
	}
	_, err := database.Exec(`PRAGMA foreign_keys=OFF;
BEGIN IMMEDIATE;
CREATE TABLE runbooks_new (
 id TEXT PRIMARY KEY, slug TEXT UNIQUE NOT NULL, title TEXT NOT NULL, body TEXT NOT NULL,
 status TEXT NOT NULL DEFAULT 'draft' CHECK(status IN ('draft','published')),
 created_by TEXT NOT NULL REFERENCES users(id), created_at TEXT NOT NULL, updated_at TEXT NOT NULL, published_at TEXT,
 folder_id TEXT REFERENCES folders(id) ON DELETE RESTRICT
);
INSERT INTO runbooks_new(id,slug,title,body,status,created_by,created_at,updated_at,published_at,folder_id)
 SELECT id,slug,title,body,CASE WHEN status='review' THEN 'draft' ELSE status END,created_by,created_at,updated_at,published_at,NULL FROM runbooks;
DROP TABLE runbooks;
ALTER TABLE runbooks_new RENAME TO runbooks;
COMMIT;
PRAGMA foreign_keys=ON;`)
	return err
}
