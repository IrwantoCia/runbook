package httpapi

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"net/http"
	"strings"
	"time"

	"runbook/internal/auth"
)

type keyHandlers struct{ db *sql.DB }

func (h keyHandlers) list(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.Query(`SELECT id,name,prefix,created_at,last_used_at,revoked_at FROM api_keys ORDER BY created_at DESC`)
	if err != nil {
		fail(w, 500, "could not list API keys")
		return
	}
	defer rows.Close()
	items := []map[string]any{}
	for rows.Next() {
		var id, name, prefix, created string
		var used, revoked sql.NullString
		if rows.Scan(&id, &name, &prefix, &created, &used, &revoked) != nil {
			fail(w, 500, "could not list API keys")
			return
		}
		items = append(items, map[string]any{"id": id, "name": name, "prefix": prefix, "created_at": created, "last_used_at": nullable(used), "revoked_at": nullable(revoked)})
	}
	jsonResponse(w, 200, map[string]any{"keys": items})
}
func (h keyHandlers) create(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name string `json:"name"`
	}
	if decode(r, &input) != nil || strings.TrimSpace(input.Name) == "" {
		fail(w, 400, "name is required")
		return
	}
	key, hash, prefix, err := auth.GenerateAPIKey()
	if err != nil {
		fail(w, 500, "could not generate API key")
		return
	}
	idbytes := make([]byte, 16)
	if _, err = rand.Read(idbytes); err != nil {
		fail(w, 500, "could not generate API key")
		return
	}
	id := hex.EncodeToString(idbytes)
	_, err = h.db.Exec(`INSERT INTO api_keys(id,name,key_hash,prefix,created_at) VALUES(?,?,?,?,?)`, id, strings.TrimSpace(input.Name), hash, prefix, time.Now().UTC().Format(time.RFC3339Nano))
	if err != nil {
		fail(w, 500, "could not create API key")
		return
	}
	jsonResponse(w, 201, map[string]any{"id": id, "name": strings.TrimSpace(input.Name), "prefix": prefix, "key": key})
}
func (h keyHandlers) revoke(w http.ResponseWriter, r *http.Request) {
	result, err := h.db.Exec(`UPDATE api_keys SET revoked_at=? WHERE id=? AND revoked_at IS NULL`, time.Now().UTC().Format(time.RFC3339Nano), r.PathValue("id"))
	if err != nil {
		fail(w, 500, "could not revoke API key")
		return
	}
	count, _ := result.RowsAffected()
	if count == 0 {
		fail(w, 404, "API key not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
func nullable(value sql.NullString) any {
	if value.Valid {
		return value.String
	}
	return nil
}
