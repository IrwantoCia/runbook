package main

import (
	"testing"

	"runbook/internal/config"
	"runbook/internal/db"
)

func TestSeedSkipsExistingUsersWithoutCredentials(t *testing.T) {
	database, err := db.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	if err := db.Migrate(database); err != nil {
		t.Fatal(err)
	}
	if _, err := database.Exec(`INSERT INTO users(id,email,password_hash,created_at) VALUES('existing','owner@example.com','hash','now')`); err != nil {
		t.Fatal(err)
	}
	if err := seed(database, config.Config{}); err != nil {
		t.Fatalf("seed existing users: %v", err)
	}
}

func TestSeedRequiresCredentialsForEmptyUsers(t *testing.T) {
	database, err := db.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	if err := db.Migrate(database); err != nil {
		t.Fatal(err)
	}
	err = seed(database, config.Config{})
	want := "no users exist; set SEED_ADMIN_EMAIL and SEED_ADMIN_PASSWORD in .env"
	if err == nil || err.Error() != want {
		t.Fatalf("seed error = %v, want %q", err, want)
	}
}
