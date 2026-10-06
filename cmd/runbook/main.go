package main

import (
	"crypto/rand"
	"database/sql"
	"embed"
	"encoding/hex"
	"fmt"
	"io/fs"
	"log"
	"log/slog"
	"net/http"
	"path"
	"strings"
	"time"

	"runbook/internal/auth"
	"runbook/internal/config"
	"runbook/internal/db"
	"runbook/internal/httpapi"
)

//go:embed all:web/dist
var frontend embed.FS

func main() {
	cfg := config.Load()
	database, err := db.Open(cfg.DBPath)
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()
	if err := db.Migrate(database); err != nil {
		log.Fatal(err)
	}
	if err := seed(database, cfg); err != nil {
		log.Fatal(err)
	}
	api := httpapi.NewRouter(database, cfg)
	assets, err := fs.Sub(frontend, "web/dist")
	if err != nil {
		log.Fatal(err)
	}
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") {
			api.ServeHTTP(w, r)
			return
		}
		name := strings.TrimPrefix(path.Clean("/"+r.URL.Path), "/")
		if name == "." || name == "" {
			name = "index.html"
		}
		if _, err := fs.Stat(assets, name); err != nil {
			name = "index.html"
		}
		http.ServeFileFS(w, r, assets, name)
	})
	server := &http.Server{Addr: cfg.Addr, Handler: httpapi.SecurityHeaders(handler, cfg.CookieSecure), ReadHeaderTimeout: 5 * time.Second}
	slog.Info("runbook listening", "addr", cfg.Addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

func seed(database *sql.DB, cfg config.Config) error {
	var count int
	if err := database.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	if cfg.SeedAdminEmail == "" || cfg.SeedAdminPassword == "" {
		return fmt.Errorf("no users exist; set SEED_ADMIN_EMAIL and SEED_ADMIN_PASSWORD in .env")
	}
	if !strings.Contains(cfg.SeedAdminEmail, "@") {
		return fmt.Errorf("SEED_ADMIN_EMAIL must contain @")
	}
	if len(cfg.SeedAdminPassword) < 8 {
		return fmt.Errorf("SEED_ADMIN_PASSWORD must be at least 8 characters")
	}
	hash, err := auth.HashPassword(cfg.SeedAdminPassword)
	if err != nil {
		return err
	}
	id, err := randomID()
	if err != nil {
		return err
	}
	_, err = database.Exec(`INSERT INTO users(id,email,password_hash,created_at) VALUES(?,?,?,?)`, id, cfg.SeedAdminEmail, hash, time.Now().UTC().Format(time.RFC3339Nano))
	return err
}

func randomID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
