package httpapi

import (
	"bytes"
	"context"
	"database/sql"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"runbook/internal/auth"
)

type contextKey string

const userContextKey contextKey = "user"

func currentUser(r *http.Request) auth.User {
	user, _ := r.Context().Value(userContextKey).(auth.User)
	return user
}

func sessionRequired(database *sql.DB, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(auth.SessionCookie)
		if err != nil {
			fail(w, 401, "authentication required")
			return
		}
		user, err := auth.LookupSession(database, cookie.Value)
		if err != nil {
			fail(w, 401, "authentication required")
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), userContextKey, user)))
	})
}

func apiKeyRequired(database *sql.DB, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Fields(r.Header.Get("Authorization"))
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			fail(w, 401, "valid API key required")
			return
		}
		hash := auth.HashToken(parts[1])
		var id string
		err := database.QueryRow(`SELECT id FROM api_keys WHERE key_hash=? AND revoked_at IS NULL`, hash).Scan(&id)
		if err != nil {
			fail(w, 401, "valid API key required")
			return
		}
		_, _ = database.Exec(`UPDATE api_keys SET last_used_at=? WHERE id=?`, time.Now().UTC().Format(time.RFC3339Nano), id)
		next.ServeHTTP(w, r)
	})
}

func bodyLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") {
			r.Body = http.MaxBytesReader(w, r.Body, 2<<20)
		}
		next.ServeHTTP(w, r)
	})
}

func SecurityHeaders(next http.Handler, secure bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := w.Header()
		header.Set("X-Content-Type-Options", "nosniff")
		header.Set("X-Frame-Options", "DENY")
		header.Set("Referrer-Policy", "no-referrer")
		header.Set("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; font-src 'self' data:; connect-src 'self'; base-uri 'none'; form-action 'self'; frame-ancestors 'none'")
		if secure {
			header.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		} else {
			header.Del("Strict-Transport-Security")
		}
		next.ServeHTTP(w, r)
	})
}

func logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		slog.Info("http request", "method", r.Method, "path", r.URL.Path, "duration", time.Since(start))
	})
}
func recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if recovered := recover(); recovered != nil {
				slog.Error("http panic", "error", recovered)
				fail(w, 500, "internal server error")
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func apiNotFound(w http.ResponseWriter, r *http.Request) { fail(w, 404, "not found") }

type responseCapture struct {
	header http.Header
	status int
	body   bytes.Buffer
}

func (c *responseCapture) Header() http.Header { return c.header }
func (c *responseCapture) WriteHeader(status int) {
	if c.status == 0 {
		c.status = status
	}
}
func (c *responseCapture) Write(body []byte) (int, error) {
	if c.status == 0 {
		c.status = http.StatusOK
	}
	return c.body.Write(body)
}

func jsonMethodErrors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/api/") {
			next.ServeHTTP(w, r)
			return
		}
		capture := &responseCapture{header: make(http.Header)}
		next.ServeHTTP(capture, r)
		if capture.status == http.StatusMethodNotAllowed {
			fail(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		for name, values := range capture.header {
			for _, value := range values {
				w.Header().Add(name, value)
			}
		}
		w.WriteHeader(capture.status)
		_, _ = w.Write(capture.body.Bytes())
	})
}
