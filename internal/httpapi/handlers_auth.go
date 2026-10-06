package httpapi

import (
	"database/sql"
	"net/http"
	"strings"
	"time"

	"runbook/internal/auth"
)

type authHandlers struct {
	db     *sql.DB
	secure bool
	ttl    time.Duration
}

func (h authHandlers) login(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if decode(r, &input) != nil {
		fail(w, 400, "invalid request")
		return
	}
	var user auth.User
	var hash string
	err := h.db.QueryRow(`SELECT id,email,password_hash FROM users WHERE email=?`, strings.ToLower(strings.TrimSpace(input.Email))).Scan(&user.ID, &user.Email, &hash)
	if err != nil {
		auth.VerifyPassword(auth.DummyHash(), input.Password)
		fail(w, 401, "invalid email or password")
		return
	}
	if !auth.VerifyPassword(hash, input.Password) {
		fail(w, 401, "invalid email or password")
		return
	}
	if err := auth.CreateSessionCookie(h.db, w, user.ID, h.ttl, h.secure); err != nil {
		fail(w, 500, "could not create session")
		return
	}
	jsonResponse(w, 200, map[string]any{"user": user})
}
func (h authHandlers) logout(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie(auth.SessionCookie); err == nil {
		_ = auth.DeleteSession(h.db, cookie.Value)
	}
	auth.ClearSessionCookie(w, h.secure)
	jsonResponse(w, 200, map[string]bool{"ok": true})
}
func (h authHandlers) me(w http.ResponseWriter, r *http.Request) {
	jsonResponse(w, 200, map[string]any{"user": currentUser(r)})
}
