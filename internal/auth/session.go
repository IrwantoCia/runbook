package auth

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"net/http"
	"time"
)

const SessionCookie = "runbook_session"

type User struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

func randomID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func SetSessionCookie(w http.ResponseWriter, token string, expires time.Time, secure bool) {
	http.SetCookie(w, &http.Cookie{Name: SessionCookie, Value: token, Path: "/", Expires: expires, HttpOnly: true, Secure: secure, SameSite: http.SameSiteLaxMode})
}

func CreateSessionCookie(database *sql.DB, w http.ResponseWriter, userID string, ttl time.Duration, secure bool) error {
	id, err := randomID()
	if err != nil {
		return err
	}
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return err
	}
	token := hex.EncodeToString(raw)
	now, expires := time.Now().UTC(), time.Now().UTC().Add(ttl)
	_, err = database.Exec(`INSERT INTO sessions(id,user_id,token_hash,expires_at,created_at) VALUES(?,?,?,?,?)`, id, userID, HashToken(token), expires.Format(time.RFC3339Nano), now.Format(time.RFC3339Nano))
	if err == nil {
		SetSessionCookie(w, token, expires, secure)
	}
	return err
}

func LookupSession(database *sql.DB, token string) (User, error) {
	var user User
	err := database.QueryRow(`SELECT u.id,u.email FROM sessions s JOIN users u ON u.id=s.user_id WHERE s.token_hash=? AND s.expires_at>?`, HashToken(token), time.Now().UTC().Format(time.RFC3339Nano)).Scan(&user.ID, &user.Email)
	return user, err
}

func DeleteSession(database *sql.DB, token string) error {
	_, err := database.Exec(`DELETE FROM sessions WHERE token_hash=?`, HashToken(token))
	return err
}

func ClearSessionCookie(w http.ResponseWriter, secure bool) {
	http.SetCookie(w, &http.Cookie{Name: SessionCookie, Value: "", Path: "/", MaxAge: -1, HttpOnly: true, Secure: secure, SameSite: http.SameSiteLaxMode})
}
