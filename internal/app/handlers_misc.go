package app

import (
	"database/sql"
	"net/http"
	"time"
)

type NewsItem struct {
	ID          int64     `json:"id"`
	Title       string    `json:"title"`
	Content     string    `json:"content"`
	PublishedAt time.Time `json:"publishedAt"`
}

type Item struct {
	ID          int64   `json:"id"`
	Name        string  `json:"name"`
	Category    string  `json:"category"`
	Price       float64 `json:"price"`
	Description string  `json:"description"`
}

type Booking struct {
	ID           int64      `json:"id"`
	ClientName   string     `json:"clientName"`
	ComputerName string     `json:"computerName"`
	StartTime    time.Time  `json:"startTime"`
	EndTime      *time.Time `json:"endTime,omitempty"`
	TariffName   string     `json:"tariffName"`
	HourlyRate   float64    `json:"hourlyRate"`
	Status       string     `json:"status"`
}

type BookingRequest struct {
	ClientName   string  `json:"clientName"`
	ComputerName string  `json:"computerName"`
	StartTime    string  `json:"startTime"`
	EndTime      *string `json:"endTime,omitempty"`
	TariffName   string  `json:"tariffName"`
	HourlyRate   float64 `json:"hourlyRate"`
}

func (a *App) listNews(w http.ResponseWriter, _ *http.Request) {
	rows, err := a.db.Query(`SELECT Id, Title, Content, PublishedAt FROM News ORDER BY PublishedAt DESC`)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	news := make([]NewsItem, 0)
	for rows.Next() {
		var item NewsItem
		var publishedRaw string
		if err := rows.Scan(&item.ID, &item.Title, &item.Content, &publishedRaw); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		publishedAt, err := time.Parse(time.RFC3339Nano, publishedRaw)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		item.PublishedAt = publishedAt
		news = append(news, item)
	}

	writeJSON(w, http.StatusOK, news)
}

func (a *App) listItems(w http.ResponseWriter, _ *http.Request) {
	rows, err := a.db.Query(`SELECT Id, Name, Category, Price, COALESCE(Description, '') FROM Items ORDER BY Id`)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	items := make([]Item, 0)
	for rows.Next() {
		var item Item
		if err := rows.Scan(&item.ID, &item.Name, &item.Category, &item.Price, &item.Description); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		items = append(items, item)
	}

	writeJSON(w, http.StatusOK, items)
}

func (a *App) listClientSessions(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	clientName, err := a.findClientNameByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			writeError(w, http.StatusNotFound, "client not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	rows, err := a.db.Query(`SELECT Id, ClientName, ComputerName, StartTime, EndTime, TotalCost, COALESCE(TariffName, ''), HourlyRate
		FROM Sessions
		WHERE ClientName = ?
		ORDER BY Id DESC`, clientName)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	sessions := make([]Session, 0)
	for rows.Next() {
		session, err := scanSession(rows)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		sessions = append(sessions, session)
	}

	writeJSON(w, http.StatusOK, sessions)
}

func (a *App) listAvailableComputers(w http.ResponseWriter, _ *http.Request) {
	rows, err := a.db.Query(`SELECT Id, Name, IsOccupied FROM Computers WHERE IsOccupied = 0 ORDER BY Id`)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	computers := make([]Computer, 0)
	for rows.Next() {
		var computer Computer
		if err := rows.Scan(&computer.ID, &computer.Name, &computer.IsOccupied); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		computer.Status = computerStatus(computer.IsOccupied)
		computers = append(computers, computer)
	}

	writeJSON(w, http.StatusOK, computers)
}

func (a *App) createBooking(w http.ResponseWriter, r *http.Request) {
	var request BookingRequest
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if err := validateBookingRequest(request); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	startTime, err := time.Parse(time.RFC3339Nano, request.StartTime)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid startTime format, expected RFC3339")
		return
	}

	var endTime *time.Time
	var endRaw any
	if request.EndTime != nil {
		parsed, err := time.Parse(time.RFC3339Nano, *request.EndTime)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid endTime format, expected RFC3339")
			return
		}
		if !parsed.After(startTime) {
			writeError(w, http.StatusBadRequest, "endTime must be after startTime")
			return
		}
		endTime = &parsed
		endRaw = parsed.Format(time.RFC3339Nano)
	}

	result, err := a.db.Exec(`INSERT INTO Bookings (ClientName, ComputerName, StartTime, EndTime, TariffName, HourlyRate, Status)
		VALUES (?, ?, ?, ?, ?, ?, ?)`, request.ClientName, request.ComputerName, startTime.Format(time.RFC3339Nano), endRaw, request.TariffName, request.HourlyRate, "booked")
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	id, _ := result.LastInsertId()
	booking := Booking{
		ID:           id,
		ClientName:   request.ClientName,
		ComputerName: request.ComputerName,
		StartTime:    startTime,
		EndTime:      endTime,
		TariffName:   request.TariffName,
		HourlyRate:   request.HourlyRate,
		Status:       "booked",
	}

	writeJSON(w, http.StatusCreated, booking)
}

func (a *App) findClientNameByID(id int64) (string, error) {
	var name string
	row := a.db.QueryRow(`SELECT Name FROM Clients WHERE Id = ?`, id)
	if err := row.Scan(&name); err != nil {
		return "", err
	}
	return name, nil
}

func validateBookingRequest(request BookingRequest) error {
	if err := requireText(request.ClientName, "clientName"); err != nil {
		return err
	}
	if err := requireText(request.ComputerName, "computerName"); err != nil {
		return err
	}
	if err := requireText(request.StartTime, "startTime"); err != nil {
		return err
	}
	if err := requireText(request.TariffName, "tariffName"); err != nil {
		return err
	}
	if request.HourlyRate <= 0 {
		return errBadRequest("hourlyRate must be greater than zero")
	}
	return nil
}

type bookingScanner interface {
	Scan(dest ...any) error
}

func (a *App) scanBooking(scanner bookingScanner) (Booking, error) {
	var booking Booking
	var startRaw string
	var endRaw sql.NullString
	if err := scanner.Scan(&booking.ID, &booking.ClientName, &booking.ComputerName, &startRaw, &endRaw, &booking.TariffName, &booking.HourlyRate, &booking.Status); err != nil {
		return Booking{}, err
	}

	startTime, err := time.Parse(time.RFC3339Nano, startRaw)
	if err != nil {
		return Booking{}, err
	}
	booking.StartTime = startTime
	if endRaw.Valid {
		parsed, err := time.Parse(time.RFC3339Nano, endRaw.String)
		if err != nil {
			return Booking{}, err
		}
		booking.EndTime = &parsed
	}
	return booking, nil
}
