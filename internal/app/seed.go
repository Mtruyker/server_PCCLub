package app

import (
	"time"

	"golang.org/x/crypto/bcrypt"
)

func (a *App) seedDemoData() error {
	var count int
	if err := a.db.QueryRow(`SELECT COUNT(*) FROM Clients`).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte("tester1"), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	clients := []Client{
		{Name: "Anatoliy", Phone: "89962659984", Email: "", Balance: 1500},
		{Name: "Ivan Ivanov", Phone: "+79991234567", Email: "ivan@mail.ru", Balance: 2300},
		{Name: "Anna Sidorova", Phone: "+79993456789", Email: "anna@yandex.ru", Balance: 850},
	}
	for _, client := range clients {
		if _, err := a.db.Exec(`INSERT INTO Clients (Name, Phone, Email, PasswordHash, Balance) VALUES (?, ?, ?, ?, ?)`,
			client.Name, client.Phone, client.Email, string(passwordHash), client.Balance); err != nil {
			return err
		}
	}

	computers := []Computer{
		{Name: "ПК-01", Zone: "standard", X: 50, Y: 150, IsOccupied: false, HourPrice: 120},
		{Name: "ПК-02", Zone: "standard", X: 150, Y: 150, IsOccupied: true, HourPrice: 120},
		{Name: "VIP-01", Zone: "vip", X: 300, Y: 150, IsOccupied: false, HourPrice: 220},
	}
	for _, computer := range computers {
		if _, err := a.db.Exec(`INSERT INTO Computers (Name, Zone, X, Y, IsOccupied, HourPrice) VALUES (?, ?, ?, ?, ?, ?)`,
			computer.Name, computer.Zone, computer.X, computer.Y, computer.IsOccupied, computer.HourPrice); err != nil {
			return err
		}
	}

	tariffs := []Tariff{
		{Name: "Базовый", CostPerHour: 120, Description: "Стандартное игровое место"},
		{Name: "VIP", CostPerHour: 220, Description: "VIP-зона с повышенным комфортом"},
	}
	for _, tariff := range tariffs {
		if _, err := a.db.Exec(`INSERT INTO Tariffs (Name, CostPerHour, Description) VALUES (?, ?, ?)`, tariff.Name, tariff.CostPerHour, tariff.Description); err != nil {
			return err
		}
	}

	news := []NewsItem{
		{Title: "Турнир в субботу", Content: "Регистрация на турнир открыта.", ImageURL: "", Date: time.Now().Add(-48 * time.Hour).UTC()},
		{Title: "Скидка на бронь", Content: "При бронировании от трех часов действует скидка.", ImageURL: "", Date: time.Now().Add(-24 * time.Hour).UTC()},
	}
	for _, item := range news {
		if _, err := a.db.Exec(`INSERT INTO News (Title, Content, ImageUrl, PublishedAt) VALUES (?, ?, ?, ?)`,
			item.Title, item.Content, item.ImageURL, item.Date.Format(time.RFC3339Nano)); err != nil {
			return err
		}
	}

	products := []Item{
		{Name: "Coca-Cola 0.5", Category: "Напитки", Price: 90, Description: "Напиток", ImageURL: ""},
		{Name: "Сэндвич", Category: "Еда", Price: 180, Description: "Быстрый перекус", ImageURL: ""},
		{Name: "Энергетик", Category: "Напитки", Price: 140, Description: "Холодный напиток", ImageURL: ""},
	}
	for _, product := range products {
		if _, err := a.db.Exec(`INSERT INTO Items (Name, Category, Price, ImageUrl, Description) VALUES (?, ?, ?, ?, ?)`,
			product.Name, product.Category, product.Price, product.ImageURL, product.Description); err != nil {
			return err
		}
	}

	now := time.Now().UTC()
	if _, err := a.db.Exec(
		`INSERT INTO Sessions (ClientId, ClientName, PcId, ComputerName, StartTime, EndTime, TotalCost, TariffName, HourlyRate) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		1, "Anatoliy", 1, "ПК-01", now.Add(-3*time.Hour).Format(time.RFC3339Nano), now.Add(-90*time.Minute).Format(time.RFC3339Nano), 180, "Базовый", 120,
	); err != nil {
		return err
	}

	_, err = a.db.Exec(
		`INSERT INTO Sessions (ClientId, ClientName, PcId, ComputerName, StartTime, EndTime, TotalCost, TariffName, HourlyRate) VALUES (?, ?, ?, ?, ?, NULL, NULL, ?, ?)`,
		2, "Ivan Ivanov", 2, "ПК-02", now.Add(-45*time.Minute).Format(time.RFC3339Nano), "Базовый", 120,
	)
	return err
}
