package app

import "time"

type Client struct {
	ID      int64   `json:"id"`
	Name    string  `json:"name"`
	Phone   string  `json:"phone"`
	Email   string  `json:"email"`
	Balance float64 `json:"balance"`
}

type Computer struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	IsOccupied bool   `json:"isOccupied"`
	Status     string `json:"status"`
}

type Tariff struct {
	ID          int64   `json:"id"`
	Name        string  `json:"name"`
	CostPerHour float64 `json:"costPerHour"`
	Description string  `json:"description"`
}

type Session struct {
	ID           int64      `json:"id"`
	ClientName   string     `json:"clientName"`
	ComputerName string     `json:"computerName"`
	StartTime    time.Time  `json:"startTime"`
	EndTime      *time.Time `json:"endTime"`
	TariffName   string     `json:"tariffName"`
	HourlyRate   *float64   `json:"hourlyRate"`
	TotalCost    *float64   `json:"totalCost"`
	Status       string     `json:"status"`
}

type StartSessionRequest struct {
	ClientName   string  `json:"clientName"`
	ComputerName string  `json:"computerName"`
	TariffName   string  `json:"tariffName"`
	HourlyRate   float64 `json:"hourlyRate"`
}

type Statistics struct {
	TotalRevenue         float64         `json:"totalRevenue"`
	CompletedSessions    int64           `json:"completedSessions"`
	ActiveSessions       int64           `json:"activeSessions"`
	DailyRevenue         []StatisticItem `json:"dailyRevenue"`
	ComputerDistribution []StatisticItem `json:"computerDistribution"`
	PopularTariffs       []StatisticItem `json:"popularTariffs"`
}

type StatisticItem struct {
	Name  string  `json:"name"`
	Value float64 `json:"value"`
}
