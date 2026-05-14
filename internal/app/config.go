package app

import "os"

type Config struct {
	Port         string
	DatabasePath string
	SeedDemoData bool
	JWTSecret    string
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

	return Config{
		Port:         port,
		DatabasePath: dbPath,
		SeedDemoData: os.Getenv("SEED_DEMO_DATA") != "false",
		JWTSecret:    jwtSecret,
	}
}
