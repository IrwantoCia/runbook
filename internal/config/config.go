package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Addr, DBPath, SeedAdminEmail, SeedAdminPassword string
	CookieSecure                                    bool
	SessionTTLHours                                 int
}

func Load() Config {
	_ = godotenv.Load()
	secure, _ := strconv.ParseBool(env("COOKIE_SECURE", "true"))
	hours, err := strconv.Atoi(env("SESSION_TTL_HOURS", "720"))
	if err != nil || hours < 1 {
		hours = 720
	}
	return Config{
		Addr: env("ADDR", "127.0.0.1:8080"), DBPath: env("DB_PATH", "./runbook.db"),
		CookieSecure: secure, SessionTTLHours: hours,
		SeedAdminEmail: env("SEED_ADMIN_EMAIL", ""), SeedAdminPassword: env("SEED_ADMIN_PASSWORD", ""),
	}
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
