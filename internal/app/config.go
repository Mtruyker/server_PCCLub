package app

import (
	"os"
	"strings"
)

type Config struct {
	Port         string
	DatabasePath string
	SeedDemoData bool
	JWTSecret    string
	AdminToken   string
}

func ConfigFromEnv() Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	dbPath := os.Getenv("DATABASE_PATH")
	if dbPath == "" {
		dbPath = "data/pcclub.db"
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "pcclub-local-dev-secret"
	}

	adminToken := strings.TrimSpace(os.Getenv("ADMIN_API_TOKEN"))
	if adminToken == "" {
		adminToken = strings.TrimSpace(os.Getenv("PCCLUB_API_TOKEN"))
	}
	if adminToken == "" {
		if raw, err := os.ReadFile("api-token.txt"); err == nil {
			adminToken = strings.TrimSpace(string(raw))
		}
	}

	return Config{
		Port:         port,
		DatabasePath: dbPath,
		SeedDemoData: os.Getenv("SEED_DEMO_DATA") != "false",
		JWTSecret:    jwtSecret,
		AdminToken:   adminToken,
	}
}
