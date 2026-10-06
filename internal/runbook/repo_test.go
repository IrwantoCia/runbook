package runbook

import (
	"database/sql"
	"errors"
	"testing"

	"runbook/internal/db"
)

func TestFolderFilteringMoveAndDirectPublish(t *testing.T) {
	database, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	database.SetMaxOpenConns(1)
	if _, err := database.Exec(`PRAGMA foreign_keys=ON`); err != nil {
		t.Fatal(err)
	}
	if err := db.Migrate(database); err != nil {
		t.Fatal(err)
	}
	if _, err := database.Exec(`INSERT INTO users(id,email,password_hash,created_at) VALUES('u1','owner@example.com','hash','now'); INSERT INTO folders(id,name,created_at) VALUES('f1','Nested','now')`); err != nil {
		t.Fatal(err)
	}
	repo := Repository{DB: database}
	service := Service{Repo: repo}
	rootBook, err := service.Create("Root document", "root body", "u1", nil)
	if err != nil {
		t.Fatal(err)
	}
	folderBook, err := service.Create("Nested document", "nested body", "u1", stringPointer("f1"))
	if err != nil {
		t.Fatal(err)
	}
	rootBooks, err := repo.List("", false, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(rootBooks) != 1 || rootBooks[0].ID != rootBook.ID {
		t.Fatalf("root listing = %+v, want only %q", rootBooks, rootBook.ID)
	}
	folderBooks, err := repo.List("", false, stringPointer("f1"))
	if err != nil {
		t.Fatal(err)
	}
	if len(folderBooks) != 1 || folderBooks[0].ID != folderBook.ID {
		t.Fatalf("folder listing = %+v, want only %q", folderBooks, folderBook.ID)
	}
	if err := service.Publish(folderBook.ID); err != nil {
		t.Fatalf("direct draft-to-published transition: %v", err)
	}
	moved, err := repo.Move(folderBook.ID, nil)
	if err != nil {
		t.Fatal(err)
	}
	if moved.FolderID != nil || moved.Slug != folderBook.Slug || moved.Status != Published {
		t.Fatalf("move changed document identity/status unexpectedly: %+v", moved)
	}
	if _, err := repo.PublicBySlug(rootBook.Slug); !errors.Is(err, ErrNotFound) {
		t.Errorf("draft public lookup err=%v, want ErrNotFound", err)
	}
	if published, err := repo.PublicBySlug(folderBook.Slug); err != nil || published.ID != folderBook.ID || published.Status != Published {
		t.Errorf("published public lookup = %+v, %v", published, err)
	}
	allPublished, err := repo.List("", true, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(allPublished) != 1 || allPublished[0].ID != folderBook.ID {
		t.Fatalf("public listing = %+v, want published document only", allPublished)
	}
}

func stringPointer(value string) *string { return &value }
