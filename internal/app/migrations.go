package app

import (
	"database/sql"
	"fmt"
)

func (a *App) migrate() error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS Clients (
			Id INTEGER PRIMARY KEY AUTOINCREMENT,
			Name TEXT NOT NULL,
			Phone TEXT,
			Email TEXT,
			PasswordHash TEXT,
			Balance REAL NOT NULL DEFAULT 0
		);`,
		`CREATE TABLE IF NOT EXISTS Computers (
			Id INTEGER PRIMARY KEY AUTOINCREMENT,
			Name TEXT NOT NULL,
			Zone TEXT NOT NULL DEFAULT 'standard',
			X INTEGER NOT NULL DEFAULT 0,
			Y INTEGER NOT NULL DEFAULT 0,
			IsOccupied INTEGER NOT NULL DEFAULT 0,
			HourPrice REAL NOT NULL DEFAULT 120
		);`,
		`CREATE TABLE IF NOT EXISTS Tariffs (
			Id INTEGER PRIMARY KEY AUTOINCREMENT,
			Name TEXT NOT NULL,
			CostPerHour REAL NOT NULL,
			Description TEXT
		);`,
		`CREATE TABLE IF NOT EXISTS Sessions (
			Id INTEGER PRIMARY KEY AUTOINCREMENT,
			ClientId INTEGER NULL,
			ClientName TEXT NOT NULL,
			PcId INTEGER NULL,
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
			ImageUrl TEXT,
			PublishedAt TEXT NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS Items (
			Id INTEGER PRIMARY KEY AUTOINCREMENT,
			Name TEXT NOT NULL,
			Category TEXT NOT NULL,
			Price REAL NOT NULL,
			ImageUrl TEXT,
			Description TEXT
		);`,
		`CREATE TABLE IF NOT EXISTS Bookings (
			Id INTEGER PRIMARY KEY AUTOINCREMENT,
			ClientId INTEGER NULL,
			ClientName TEXT NOT NULL,
			PcId INTEGER NULL,
			ComputerName TEXT NOT NULL,
			StartTime TEXT NOT NULL,
			EndTime TEXT NULL,
			TariffName TEXT,
			HourlyRate REAL NOT NULL,
			DurationHours INTEGER NOT NULL DEFAULT 0,
			TotalPrice REAL NOT NULL DEFAULT 0,
			Status TEXT NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS Orders (
			Id INTEGER PRIMARY KEY AUTOINCREMENT,
			ClientId INTEGER NOT NULL,
			Date TEXT NOT NULL,
			TotalAmount REAL NOT NULL,
			Status TEXT NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS OrderItems (
			Id INTEGER PRIMARY KEY AUTOINCREMENT,
			OrderId INTEGER NOT NULL,
			ProductId INTEGER NOT NULL,
			ProductName TEXT NOT NULL,
			Quantity INTEGER NOT NULL,
			Price REAL NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS OrderStatusHistory (
			Id INTEGER PRIMARY KEY AUTOINCREMENT,
			OrderId INTEGER NOT NULL,
			Status TEXT NOT NULL,
			ChangedAt TEXT NOT NULL
		);`,
	}

	for _, statement := range statements {
		if _, err := a.db.Exec(statement); err != nil {
			return fmt.Errorf("migrate: %w", err)
		}
	}

	columns := []struct {
		table      string
		name       string
		definition string
	}{
		{"Clients", "PasswordHash", "TEXT"},
		{"Computers", "Zone", "TEXT NOT NULL DEFAULT 'standard'"},
		{"Computers", "X", "INTEGER NOT NULL DEFAULT 0"},
		{"Computers", "Y", "INTEGER NOT NULL DEFAULT 0"},
		{"Computers", "HourPrice", "REAL NOT NULL DEFAULT 120"},
		{"Sessions", "ClientId", "INTEGER NULL"},
		{"Sessions", "PcId", "INTEGER NULL"},
		{"News", "ImageUrl", "TEXT"},
		{"Items", "ImageUrl", "TEXT"},
		{"Bookings", "ClientId", "INTEGER NULL"},
		{"Bookings", "PcId", "INTEGER NULL"},
		{"Bookings", "DurationHours", "INTEGER NOT NULL DEFAULT 0"},
		{"Bookings", "TotalPrice", "REAL NOT NULL DEFAULT 0"},
	}
	for _, column := range columns {
		if err := a.ensureColumn(column.table, column.name, column.definition); err != nil {
			return fmt.Errorf("migrate %s.%s: %w", column.table, column.name, err)
		}
	}

	indexes := []string{
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_clients_phone ON Clients(Phone) WHERE Phone IS NOT NULL AND Phone <> '';`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_clients_email ON Clients(Email) WHERE Email IS NOT NULL AND Email <> '';`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_computers_name ON Computers(Name);`,
		`CREATE INDEX IF NOT EXISTS idx_bookings_pc_time ON Bookings(PcId, StartTime, Status);`,
		`CREATE INDEX IF NOT EXISTS idx_orders_client ON Orders(ClientId);`,
		`CREATE INDEX IF NOT EXISTS idx_order_status_history_order ON OrderStatusHistory(OrderId);`,
		`CREATE INDEX IF NOT EXISTS idx_sessions_client ON Sessions(ClientId);`,
	}
	for _, statement := range indexes {
		if _, err := a.db.Exec(statement); err != nil {
			return fmt.Errorf("migrate indexes: %w", err)
		}
	}

	return nil
}

func (a *App) ensureColumn(table, name, definition string) error {
	rows, err := a.db.Query(`PRAGMA table_info(` + table + `)`)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var cid int
		var columnName, columnType string
		var notNull int
		var defaultValue sql.NullString
		var pk int
		if err := rows.Scan(&cid, &columnName, &columnType, &notNull, &defaultValue, &pk); err != nil {
			return err
		}
		if columnName == name {
			return nil
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}

	_, err = a.db.Exec(`ALTER TABLE ` + table + ` ADD COLUMN ` + name + ` ` + definition)
	return err
}
