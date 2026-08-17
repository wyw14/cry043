package config

import (
	"os"
	"time"
)

type Config struct {
	Addr, DatabaseURL, UploadDir string
	Timeout                      time.Duration
}

func Load() Config {
	return Config{get("HTTP_ADDR", ":8080"), get("DATABASE_URL", "postgres://spec:spec@localhost:5432/spec?sslmode=disable"), get("UPLOAD_DIR", "./var/evidence"), 5 * time.Second}
}
func get(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
