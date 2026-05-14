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
		`CREATE TABLE IF NOT EXISTS News (
			Id INTEGER PRIMARY KEY AUTOINCREMENT,
			Title TEXT NOT NULL,
			Content TEXT NOT NULL,
			PublishedAt TEXT NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS Items (
			Id INTEGER PRIMARY KEY AUTOINCREMENT,
			Name TEXT NOT NULL,
			Category TEXT NOT NULL,
			Price REAL NOT NULL,
			Description TEXT
		);`,
		`CREATE TABLE IF NOT EXISTS Bookings (
			Id INTEGER PRIMARY KEY AUTOINCREMENT,
			ClientName TEXT NOT NULL,
			ComputerName TEXT NOT NULL,
			StartTime TEXT NOT NULL,
			EndTime TEXT NULL,
			TariffName TEXT,
			HourlyRate REAL NOT NULL,
			Status TEXT NOT NULL
		);`,
	}

	for _, statement := range statements {
		if _, err := a.db.Exec(statement); err != nil {
			return fmt.Errorf("migrate: %w", err)
		}
	}

	return nil
}
