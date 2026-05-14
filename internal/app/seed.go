package app

import "time"

func (a *App) seedDemoData() error {
	var count int
	if err := a.db.QueryRow(`SELECT COUNT(*) FROM Clients`).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	clients := []Client{
		{Name: "Иванов Иван", Phone: "+7 (999) 123-45-67", Email: "ivan@mail.ru", Balance: 1500},
		{Name: "Петров Петр", Phone: "+7 (999) 234-56-78", Email: "petrov@gmail.com", Balance: 2300},
		{Name: "Сидорова Анна", Phone: "+7 (999) 345-67-89", Email: "anna@yandex.ru", Balance: 850},
	}
	for _, client := range clients {
		if _, err := a.db.Exec(`INSERT INTO Clients (Name, Phone, Email, Balance) VALUES (?, ?, ?, ?)`, client.Name, client.Phone, client.Email, client.Balance); err != nil {
			return err
		}
	}

	computers := []Computer{
		{Name: "ПК-01", IsOccupied: false},
		{Name: "ПК-02", IsOccupied: true},
		{Name: "VIP-01", IsOccupied: false},
	}
	for _, computer := range computers {
		if _, err := a.db.Exec(`INSERT INTO Computers (Name, IsOccupied) VALUES (?, ?)`, computer.Name, computer.IsOccupied); err != nil {
			return err
		}
	}

	tariffs := []Tariff{
		{Name: "Базовый", CostPerHour: 300, Description: "Стандартный тариф"},
		{Name: "Премиум", CostPerHour: 450, Description: "Повышенная скорость и приоритет"},
		{Name: "VIP", CostPerHour: 650, Description: "Игровое место повышенного класса"},
	}
	for _, tariff := range tariffs {
		if _, err := a.db.Exec(`INSERT INTO Tariffs (Name, CostPerHour, Description) VALUES (?, ?, ?)`, tariff.Name, tariff.CostPerHour, tariff.Description); err != nil {
			return err
		}
	}

	news := []NewsItem{
		{Title: "Открытие новых зон", Content: "Скоро будет открыта ещё одна игровая зона с современными ПК.", PublishedAt: time.Now().Add(-48 * time.Hour).UTC()},
		{Title: "Акция на аренду", Content: "При бронировании более 3 часов действует скидка 10%.", PublishedAt: time.Now().Add(-24 * time.Hour).UTC()},
	}
	for _, item := range news {
		if _, err := a.db.Exec(`INSERT INTO News (Title, Content, PublishedAt) VALUES (?, ?, ?)`, item.Title, item.Content, item.PublishedAt.Format(time.RFC3339Nano)); err != nil {
			return err
		}
	}

	products := []Item{
		{Name: "Игровая мышь", Category: "Аксессуары", Price: 1250, Description: "Высокоточная мышь для профессионалов."},
		{Name: "Игровая клавиатура", Category: "Аксессуары", Price: 2100, Description: "Механическая клавиатура с RGB-подсветкой."},
		{Name: "Наушники", Category: "Аудиотехника", Price: 1750, Description: "Комфортные наушники с шумоподавлением."},
	}
	for _, product := range products {
		if _, err := a.db.Exec(`INSERT INTO Items (Name, Category, Price, Description) VALUES (?, ?, ?, ?)`, product.Name, product.Category, product.Price, product.Description); err != nil {
			return err
		}
	}

	now := time.Now().UTC()
	if _, err := a.db.Exec(
		`INSERT INTO Sessions (ClientName, ComputerName, StartTime, EndTime, TotalCost, TariffName, HourlyRate) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		"Иванов Иван", "ПК-01", now.Add(-3*time.Hour).Format(time.RFC3339Nano), now.Add(-90*time.Minute).Format(time.RFC3339Nano), 450, "Базовый", 300,
	); err != nil {
		return err
	}

	_, err := a.db.Exec(
		`INSERT INTO Sessions (ClientName, ComputerName, StartTime, EndTime, TotalCost, TariffName, HourlyRate) VALUES (?, ?, ?, NULL, NULL, ?, ?)`,
		"Петров Петр", "ПК-02", now.Add(-45*time.Minute).Format(time.RFC3339Nano), "Базовый", 300,
	)
	return err
}
