package app

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

type App struct {
	cfg Config
	db  *sql.DB
}

func New(cfg Config) (*App, error) {
	if err := os.MkdirAll(filepath.Dir(cfg.DatabasePath), 0o755); err != nil {
		return nil, fmt.Errorf("create data directory: %w", err)
	}

	db, err := sql.Open("sqlite", cfg.DatabasePath)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	app := &App{cfg: cfg, db: db}
	if err := app.migrate(); err != nil {
		db.Close()
		return nil, err
	}

	if cfg.SeedDemoData {
		if err := app.seedDemoData(); err != nil {
			db.Close()
			return nil, err
		}
	}

	return app, nil
}

func (a *App) Close() error {
	return a.db.Close()
}

func (a *App) Routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", a.health)
	mux.HandleFunc("GET /api/clients", a.listClients)
	mux.HandleFunc("POST /api/clients", a.createClient)
	mux.HandleFunc("PUT /api/clients/{id}", a.updateClient)
	mux.HandleFunc("DELETE /api/clients/{id}", a.deleteClient)

	mux.HandleFunc("GET /api/computers", a.listComputers)
	mux.HandleFunc("POST /api/computers", a.createComputer)
	mux.HandleFunc("PUT /api/computers/{id}", a.updateComputer)
	mux.HandleFunc("DELETE /api/computers/{id}", a.deleteComputer)

	mux.HandleFunc("GET /api/tariffs", a.listTariffs)
	mux.HandleFunc("POST /api/tariffs", a.createTariff)
	mux.HandleFunc("PUT /api/tariffs/{id}", a.updateTariff)
	mux.HandleFunc("DELETE /api/tariffs/{id}", a.deleteTariff)

	mux.HandleFunc("GET /api/sessions", a.listSessions)
	mux.HandleFunc("POST /api/sessions", a.startSession)
	mux.HandleFunc("POST /api/sessions/{id}/complete", a.completeSession)
	mux.HandleFunc("DELETE /api/sessions/{id}", a.deleteSession)

	mux.HandleFunc("GET /api/statistics", a.statistics)

	return cors(mux)
}

func (a *App) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
