package folder

import (
	"database/sql"
	"errors"
	"testing"

	"runbook/internal/db"
)

func TestHierarchyValidationAndEmptyOnlyDelete(t *testing.T) {
	database, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	database.SetMaxOpenConns(1)
	if err := db.Migrate(database); err != nil {
		t.Fatal(err)
	}
	if _, err := database.Exec(`INSERT INTO users(id,email,password_hash,created_at) VALUES('u1','owner@example.com','hash','now')`); err != nil {
		t.Fatal(err)
	}
	repo := Repository{DB: database}
	root, err := repo.Create("root", "Root", nil)
	if err != nil {
		t.Fatal(err)
	}
	child, err := repo.Create("child", "Child", &root.ID)
	if err != nil {
		t.Fatal(err)
	}
	grandchild, err := repo.Create("grandchild", "Grandchild", &child.ID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := repo.Create("duplicate-root", "ROOT", nil); !errors.Is(err, ErrDuplicate) {
		t.Errorf("duplicate root sibling err=%v, want ErrDuplicate", err)
	}
	if _, err := repo.Create("duplicate-child", "CHILD", &root.ID); !errors.Is(err, ErrDuplicate) {
		t.Errorf("duplicate nested sibling err=%v, want ErrDuplicate", err)
	}
	if _, err := repo.Create("missing-parent", "Missing parent", stringPointer("missing")); !errors.Is(err, ErrNotFound) {
		t.Errorf("missing parent create err=%v, want ErrNotFound", err)
	}
	if err := repo.Move(root.ID, &root.ID); !errors.Is(err, ErrCycle) {
		t.Errorf("self move err=%v, want ErrCycle", err)
	}
	if err := repo.Move(root.ID, &grandchild.ID); !errors.Is(err, ErrCycle) {
		t.Errorf("descendant move err=%v, want ErrCycle", err)
	}
	if err := repo.Move(child.ID, stringPointer("missing")); !errors.Is(err, ErrNotFound) {
		t.Errorf("missing move parent err=%v, want ErrNotFound", err)
	}
	rootSibling, err := repo.Create("root-sibling", "Other root", nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := repo.Rename(rootSibling.ID, "ROOT"); !errors.Is(err, ErrDuplicate) {
		t.Errorf("duplicate root rename err=%v, want ErrDuplicate", err)
	}
	if err := repo.Delete(root.ID); !errors.Is(err, ErrNotEmpty) {
		t.Errorf("nonempty folder delete err=%v, want ErrNotEmpty", err)
	}
	if err := repo.Delete(grandchild.ID); err != nil {
		t.Fatal(err)
	}
	if err := repo.Delete(child.ID); err != nil {
		t.Fatal(err)
	}
	if err := repo.Delete(rootSibling.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := database.Exec(`INSERT INTO runbooks(id,slug,title,body,status,created_by,created_at,updated_at,folder_id) VALUES('doc','doc','Doc','body','draft','u1','now','now',?)`, root.ID); err != nil {
		t.Fatal(err)
	}
	if err := repo.Delete(root.ID); !errors.Is(err, ErrNotEmpty) {
		t.Errorf("folder with document delete err=%v, want ErrNotEmpty", err)
	}
}

func stringPointer(value string) *string { return &value }
