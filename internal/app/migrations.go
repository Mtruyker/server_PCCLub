package app

import "fmt"

func (a *App) migrate() error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS Clients (
			Id INTEGER PRIMARY KEY AUTOINCREMENT,
			Name TEXT NOT NULL,
			Phone TEXT,
			Email TEXT,
			Balance REAL NOT NULL DEFAULT 0
		);`,
		`CREATE TABLE IF NOT EXISTS Computers (
			Id INTEGER PRIMARY KEY AUTOINCREMENT,
			Name TEXT NOT NULL,
			IsOccupied INTEGER NOT NULL DEFAULT 0
		);`,
		`CREATE TABLE IF NOT EXISTS Tariffs (
			Id INTEGER PRIMARY KEY AUTOINCREMENT,
			Name TEXT NOT NULL,
			CostPerHour REAL NOT NULL,
			Description TEXT
		);`,
		`CREATE TABLE IF NOT EXISTS Sessions (
			Id INTEGER PRIMARY KEY AUTOINCREMENT,
			ClientName TEXT NOT NULL,
			ComputerName TEXT NOT NULL,
			StartTime TEXT NOT NULL,
			EndTime TEXT NULL,
			TotalCost REAL NULL,
			TariffName TEXT NULL,
			HourlyRate REAL NULL
		);`,
	}

	for _, statement := range statements {
		if _, err := a.db.Exec(statement); err != nil {
			return fmt.Errorf("migrate: %w", err)
		}
	}

	return nil
}
