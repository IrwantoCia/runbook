package db

import (
	"database/sql"
	"fmt"
	"strings"

	_ "modernc.org/sqlite"
)

func Open(path string) (*sql.DB, error) {
	dsn := path
	if !strings.Contains(path, "?") {
		dsn += "?"
	} else {
		dsn += "&"
	}
	dsn += "_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)"
	if strings.HasPrefix(path, "file:") {
		dsn = path + "&_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)"
	}
	database, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	database.SetMaxOpenConns(1)
	if err := database.Ping(); err != nil {
		database.Close()
		return nil, fmt.Errorf("ping sqlite: %w", err)
	}
	return database, nil
}
