package app

import "os"

type Config struct {
	Port         string
	DatabasePath string
	SeedDemoData bool
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

	return Config{
		Port:         port,
		DatabasePath: dbPath,
		SeedDemoData: os.Getenv("SEED_DEMO_DATA") != "false",
	}
}
