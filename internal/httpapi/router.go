package httpapi

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"net/http"
	"time"

	"runbook/internal/config"
	"runbook/internal/folder"
	"runbook/internal/runbook"
)

func NewRouter(database *sql.DB, cfg config.Config) http.Handler {
	mux := http.NewServeMux()
	authAPI := authHandlers{db: database, secure: cfg.CookieSecure, ttl: time.Duration(cfg.SessionTTLHours) * time.Hour}
	books := runbookHandlers{repo: runbook.Repository{DB: database}}
	books.service = runbook.Service{Repo: books.repo}
	folders := folderHandlers{repo: folder.Repository{DB: database}}
	keys := keyHandlers{db: database}
	public := publicHandlers{repo: books.repo}
	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, r *http.Request) { jsonResponse(w, 200, map[string]string{"status": "ok"}) })
	mux.HandleFunc("POST /api/auth/login", authAPI.login)
	mux.Handle("POST /api/auth/logout", sessionRequired(database, http.HandlerFunc(authAPI.logout)))
	mux.Handle("GET /api/auth/me", sessionRequired(database, http.HandlerFunc(authAPI.me)))
	mux.Handle("GET /api/runbooks", sessionRequired(database, http.HandlerFunc(books.list)))
	mux.Handle("POST /api/runbooks", sessionRequired(database, http.HandlerFunc(books.create)))
	mux.Handle("POST /api/runbooks/{id}/move", sessionRequired(database, http.HandlerFunc(books.move)))
	mux.Handle("GET /api/runbooks/{id}", sessionRequired(database, http.HandlerFunc(books.get)))
	mux.Handle("PATCH /api/runbooks/{id}", sessionRequired(database, http.HandlerFunc(books.update)))
	mux.Handle("DELETE /api/runbooks/{id}", sessionRequired(database, http.HandlerFunc(books.delete)))
	mux.Handle("GET /api/folders", sessionRequired(database, http.HandlerFunc(folders.list)))
	mux.Handle("POST /api/folders", sessionRequired(database, http.HandlerFunc(folders.create)))
	mux.Handle("PATCH /api/folders/{id}", sessionRequired(database, http.HandlerFunc(folders.rename)))
	mux.Handle("POST /api/folders/{id}/move", sessionRequired(database, http.HandlerFunc(folders.move)))
	mux.Handle("DELETE /api/folders/{id}", sessionRequired(database, http.HandlerFunc(folders.delete)))
	mux.Handle("POST /api/runbooks/{id}/publish", sessionRequired(database, http.HandlerFunc(books.publish)))
	mux.Handle("GET /api/keys", sessionRequired(database, http.HandlerFunc(keys.list)))
	mux.Handle("POST /api/keys", sessionRequired(database, http.HandlerFunc(keys.create)))
	mux.Handle("DELETE /api/keys/{id}", sessionRequired(database, http.HandlerFunc(keys.revoke)))
	mux.Handle("GET /api/v1/runbooks", apiKeyRequired(database, http.HandlerFunc(public.list)))
	mux.Handle("GET /api/v1/runbooks/{slug}", apiKeyRequired(database, http.HandlerFunc(public.get)))
	mux.HandleFunc("/api/", apiNotFound)
	return SecurityHeaders(recoverer(logging(jsonMethodErrors(bodyLimit(mux)))), cfg.CookieSecure)
}

func newFolderID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
