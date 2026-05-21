package app

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Port            string
	DatabasePath    string
	SeedDemoData    bool
	JWTSecret       string
	AdminToken      string

	// Server performance settings
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	MaxHeaderBytes  int

	// Database settings
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
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

	// Parse performance settings with defaults
	readTimeout := parseDurationEnv("READ_TIMEOUT", 15*time.Second)
	writeTimeout := parseDurationEnv("WRITE_TIMEOUT", 15*time.Second)
	idleTimeout := parseDurationEnv("IDLE_TIMEOUT", 60*time.Second)
	maxHeaderBytes := parseIntEnv("MAX_HEADER_BYTES", 1<<20) // 1MB

	// Database connection pool settings
	maxOpenConns := parseIntEnv("DB_MAX_OPEN_CONNS", 25)
	maxIdleConns := parseIntEnv("DB_MAX_IDLE_CONNS", 5)
	connMaxLifetime := parseDurationEnv("DB_CONN_MAX_LIFETIME", 5*time.Minute)
	connMaxIdleTime := parseDurationEnv("DB_CONN_MAX_IDLE_TIME", 5*time.Minute)

	return Config{
		Port:         port,
		DatabasePath: dbPath,
		SeedDemoData: os.Getenv("SEED_DEMO_DATA") != "false",
		JWTSecret:    jwtSecret,
		AdminToken:   adminToken,

		ReadTimeout:     readTimeout,
		WriteTimeout:    writeTimeout,
		IdleTimeout:     idleTimeout,
		MaxHeaderBytes:  maxHeaderBytes,

		MaxOpenConns:    maxOpenConns,
		MaxIdleConns:    maxIdleConns,
		ConnMaxLifetime: connMaxLifetime,
		ConnMaxIdleTime: connMaxIdleTime,
	}
}

func parseIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func parseDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if parsed, err := time.ParseDuration(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}
