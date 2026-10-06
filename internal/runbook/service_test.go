package runbook

import (
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"
)

func TestSlugify(t *testing.T) {
	for _, test := range []struct{ title, want string }{{" Restore API: Part 2! ", "restore-api-part-2"}, {"東京", "runbook"}} {
		if got := Slugify(test.title); got != test.want {
			t.Errorf("Slugify(%q) = %q, want %q", test.title, got, test.want)
		}
	}
}

func TestLifecycle(t *testing.T) {
	database, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	_, err = database.Exec(`CREATE TABLE users(id TEXT PRIMARY KEY,email TEXT); CREATE TABLE folders(id TEXT PRIMARY KEY,name TEXT,parent_id TEXT,created_at TEXT); CREATE TABLE runbooks(id TEXT PRIMARY KEY,slug TEXT UNIQUE,title TEXT,body TEXT,status TEXT,created_by TEXT,created_at TEXT,updated_at TEXT,published_at TEXT,folder_id TEXT); INSERT INTO users VALUES('user-1','test@example.com')`)
	if err != nil {
		t.Fatal(err)
	}
	service := Service{Repo: Repository{DB: database}}
	book, err := service.Create("Deploy API", "steps", "user-1", nil)
	if err != nil {
		t.Fatal(err)
	}
	if book.Status != Draft || book.Slug != "deploy-api" {
		t.Fatalf("unexpected initial runbook: %+v", book)
	}
	if err := service.Publish(book.ID); err != nil {
		t.Fatal(err)
	}
	book, err = service.Repo.Get(book.ID)
	if err != nil {
		t.Fatal(err)
	}
	if book.Status != Published || book.PublishedAt == nil {
		t.Fatalf("runbook was not published: %+v", book)
	}
	publicBook, err := service.Repo.PublicBySlug(book.Slug)
	if err != nil || publicBook.ID != book.ID {
		t.Fatalf("public lookup failed: book=%+v err=%v", publicBook, err)
	}
	if err := service.Publish(book.ID); err == nil {
		t.Fatal("expected invalid published -> published transition")
	}
}
