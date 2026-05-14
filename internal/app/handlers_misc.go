package app

import (
	"database/sql"
	"net/http"
	"time"
)

type NewsItem struct {
	ID       int64     `json:"id"`
	Title    string    `json:"title"`
	Content  string    `json:"content"`
	ImageURL string    `json:"imageUrl"`
	Date     time.Time `json:"date"`
}

type Item struct {
	ID          int64   `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	ImageURL    string  `json:"imageUrl"`
	Category    string  `json:"category"`
}

type AvailablePC struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	IsOccupied bool   `json:"isOccupied"`
	Status     string `json:"status"`
}

type Booking struct {
	ID            int64     `json:"id"`
	ClientID      int64     `json:"clientId"`
	PCName        string    `json:"pcName"`
	StartTime     time.Time `json:"startTime"`
	DurationHours int       `json:"durationHours"`
	TotalPrice    float64   `json:"totalPrice"`
	Status        string    `json:"status"`
}

type BookingRequest struct {
	ClientID     int64  `json:"clientId"`
	PCName       string `json:"pcName"`
	ComputerName string `json:"computerName"`
	StartTime    string `json:"startTime"`
	Duration     int    `json:"duration"`
}

type ClientSession struct {
	ID        int64      `json:"id"`
	PCName    string     `json:"pcName"`
	StartTime time.Time  `json:"startTime"`
	EndTime   *time.Time `json:"endTime"`
	Cost      float64    `json:"cost"`
}

type OrderItemRequest struct {
	ProductID int64 `json:"productId"`
	Quantity  int   `json:"quantity"`
}

type OrderRequest struct {
	ClientID int64              `json:"clientId"`
	Items    []OrderItemRequest `json:"items"`
	Date     string             `json:"date"`
}

type OrderItemResponse struct {
	ProductName string  `json:"productName"`
	Quantity    int     `json:"quantity"`
	Price       float64 `json:"price"`
}

type Order struct {
	ID          int64               `json:"id"`
	ClientID    int64               `json:"clientId"`
	Date        time.Time           `json:"date"`
	Items       []OrderItemResponse `json:"items"`
	TotalAmount float64             `json:"totalAmount"`
	Status      string              `json:"status"`
}

func (a *App) listNews(w http.ResponseWriter, _ *http.Request) {
	rows, err := a.db.Query(`SELECT Id, Title, Content, COALESCE(ImageUrl, ''), PublishedAt FROM News ORDER BY PublishedAt DESC`)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	news := make([]NewsItem, 0)
	for rows.Next() {
		var item NewsItem
		var dateRaw string
		if err := rows.Scan(&item.ID, &item.Title, &item.Content, &item.ImageURL, &dateRaw); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		date, err := time.Parse(time.RFC3339Nano, dateRaw)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		item.Date = date
		news = append(news, item)
	}

	writeJSON(w, http.StatusOK, news)
}

func (a *App) listItems(w http.ResponseWriter, _ *http.Request) {
	rows, err := a.db.Query(`SELECT Id, Name, COALESCE(Description, ''), Price, COALESCE(ImageUrl, ''), Category FROM Items ORDER BY Id`)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	items := make([]Item, 0)
	for rows.Next() {
		var item Item
		if err := rows.Scan(&item.ID, &item.Name, &item.Description, &item.Price, &item.ImageURL, &item.Category); err != nil {
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
	if err := a.requireClient(r, id); err != nil {
		writeAuthError(w, err)
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

	rows, err := a.db.Query(`SELECT Id, ComputerName, StartTime, EndTime, COALESCE(TotalCost, 0)
		FROM Sessions
		WHERE ClientId = ? OR ClientName = ?
		ORDER BY Id DESC`, id, clientName)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	sessions := make([]ClientSession, 0)
	for rows.Next() {
		var session ClientSession
		var startRaw string
		var endRaw sql.NullString
		if err := rows.Scan(&session.ID, &session.PCName, &startRaw, &endRaw, &session.Cost); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		session.StartTime, err = time.Parse(time.RFC3339Nano, startRaw)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if endRaw.Valid {
			endTime, err := time.Parse(time.RFC3339Nano, endRaw.String)
			if err != nil {
				writeError(w, http.StatusInternalServerError, err.Error())
				return
			}
			session.EndTime = &endTime
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

	computers := make([]AvailablePC, 0)
	for rows.Next() {
		var computer AvailablePC
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
	if request.PCName == "" {
		request.PCName = request.ComputerName
	}
	if err := validateBookingRequest(request); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := a.requireClient(r, request.ClientID); err != nil {
		writeAuthError(w, err)
		return
	}

	client, err := a.findClientByID(request.ClientID)
	if err != nil {
		if err == sql.ErrNoRows {
			writeError(w, http.StatusNotFound, "client not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	pcID, hourPrice, err := a.findPCForBooking(request.PCName)
	if err != nil {
		if err == sql.ErrNoRows {
			writeError(w, http.StatusNotFound, "pc not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	startTime, err := time.Parse(time.RFC3339Nano, request.StartTime)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid startTime format, expected RFC3339")
		return
	}
	endTime := startTime.Add(time.Duration(request.Duration) * time.Hour)
	if err := a.ensureBookingSlotAvailable(pcID, request.PCName, startTime, endTime); err != nil {
		writeError(w, http.StatusConflict, err.Error())
		return
	}

	totalPrice := hourPrice * float64(request.Duration)
	result, err := a.db.Exec(`INSERT INTO Bookings
		(ClientId, ClientName, PcId, ComputerName, StartTime, EndTime, TariffName, HourlyRate, DurationHours, TotalPrice, Status)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		request.ClientID, client.Name, pcID, request.PCName, startTime.Format(time.RFC3339Nano), endTime.Format(time.RFC3339Nano),
		"", hourPrice, request.Duration, totalPrice, "active",
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	id, _ := result.LastInsertId()
	writeJSON(w, http.StatusCreated, Booking{
		ID:            id,
		ClientID:      request.ClientID,
		PCName:        request.PCName,
		StartTime:     startTime,
		DurationHours: request.Duration,
		TotalPrice:    totalPrice,
		Status:        "active",
	})
}

func (a *App) listClientBookings(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := a.requireClient(r, id); err != nil {
		writeAuthError(w, err)
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

	rows, err := a.db.Query(`SELECT Id, COALESCE(ClientId, 0), ComputerName, StartTime, EndTime, COALESCE(DurationHours, 0), COALESCE(TotalPrice, 0), COALESCE(HourlyRate, 0), Status
		FROM Bookings
		WHERE ClientId = ? OR ClientName = ?
		ORDER BY StartTime DESC`, id, clientName)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	bookings := make([]Booking, 0)
	for rows.Next() {
		booking, err := scanMobileBooking(rows)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if booking.ClientID == 0 {
			booking.ClientID = id
		}
		bookings = append(bookings, booking)
	}

	writeJSON(w, http.StatusOK, bookings)
}

func (a *App) cancelBooking(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	var clientID sql.NullInt64
	var clientName string
	err = a.db.QueryRow(`SELECT ClientId, ClientName FROM Bookings WHERE Id = ?`, id).Scan(&clientID, &clientName)
	if err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "booking not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !clientID.Valid {
		var resolvedID int64
		if err := a.db.QueryRow(`SELECT Id FROM Clients WHERE Name = ? ORDER BY Id LIMIT 1`, clientName).Scan(&resolvedID); err == nil {
			clientID = sql.NullInt64{Int64: resolvedID, Valid: true}
		}
	}
	if !clientID.Valid {
		writeError(w, http.StatusForbidden, "booking owner is unknown")
		return
	}
	if err := a.requireClient(r, clientID.Int64); err != nil {
		writeAuthError(w, err)
		return
	}

	if _, err := a.db.Exec(`UPDATE Bookings SET Status = 'cancelled' WHERE Id = ?`, id); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *App) createOrder(w http.ResponseWriter, r *http.Request) {
	var request OrderRequest
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if err := validateOrderRequest(request); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := a.requireClient(r, request.ClientID); err != nil {
		writeAuthError(w, err)
		return
	}

	orderDate := time.Now().UTC()
	if request.Date != "" {
		parsed, err := time.Parse(time.RFC3339Nano, request.Date)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid date format, expected RFC3339")
			return
		}
		orderDate = parsed
	}

	tx, err := a.db.Begin()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer tx.Rollback()

	items := make([]OrderItemResponse, 0, len(request.Items))
	totalAmount := 0.0
	for _, requestedItem := range request.Items {
		var name string
		var price float64
		err := tx.QueryRow(`SELECT Name, Price FROM Items WHERE Id = ?`, requestedItem.ProductID).Scan(&name, &price)
		if err == sql.ErrNoRows {
			writeError(w, http.StatusNotFound, "product not found")
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		items = append(items, OrderItemResponse{
			ProductName: name,
			Quantity:    requestedItem.Quantity,
			Price:       price,
		})
		totalAmount += price * float64(requestedItem.Quantity)
	}

	result, err := tx.Exec(`INSERT INTO Orders (ClientId, Date, TotalAmount, Status) VALUES (?, ?, ?, ?)`,
		request.ClientID, orderDate.Format(time.RFC3339Nano), totalAmount, "pending")
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	orderID, _ := result.LastInsertId()

	for i, requestedItem := range request.Items {
		item := items[i]
		if _, err := tx.Exec(`INSERT INTO OrderItems (OrderId, ProductId, ProductName, Quantity, Price) VALUES (?, ?, ?, ?, ?)`,
			orderID, requestedItem.ProductID, item.ProductName, item.Quantity, item.Price); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}

	if err := tx.Commit(); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, Order{
		ID:          orderID,
		ClientID:    request.ClientID,
		Date:        orderDate,
		Items:       items,
		TotalAmount: totalAmount,
		Status:      "pending",
	})
}

func (a *App) listClientOrders(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := a.requireClient(r, id); err != nil {
		writeAuthError(w, err)
		return
	}

	rows, err := a.db.Query(`SELECT Id, ClientId, Date, TotalAmount, Status FROM Orders WHERE ClientId = ? ORDER BY Date DESC`, id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	orders := make([]Order, 0)
	for rows.Next() {
		var order Order
		var dateRaw string
		if err := rows.Scan(&order.ID, &order.ClientID, &dateRaw, &order.TotalAmount, &order.Status); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		order.Date, err = time.Parse(time.RFC3339Nano, dateRaw)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		order.Items, err = a.listOrderItems(order.ID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		orders = append(orders, order)
	}

	writeJSON(w, http.StatusOK, orders)
}

func (a *App) listOrderItems(orderID int64) ([]OrderItemResponse, error) {
	rows, err := a.db.Query(`SELECT ProductName, Quantity, Price FROM OrderItems WHERE OrderId = ? ORDER BY Id`, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]OrderItemResponse, 0)
	for rows.Next() {
		var item OrderItemResponse
		if err := rows.Scan(&item.ProductName, &item.Quantity, &item.Price); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (a *App) findClientNameByID(id int64) (string, error) {
	var name string
	row := a.db.QueryRow(`SELECT Name FROM Clients WHERE Id = ?`, id)
	if err := row.Scan(&name); err != nil {
		return "", err
	}
	return name, nil
}

func (a *App) findPCForBooking(pcName string) (int64, float64, error) {
	var id int64
	var hourPrice float64
	err := a.db.QueryRow(`SELECT Id, COALESCE(HourPrice, 120) FROM Computers WHERE Name = ?`, pcName).Scan(&id, &hourPrice)
	return id, hourPrice, err
}

func (a *App) ensureBookingSlotAvailable(pcID int64, pcName string, startTime, endTime time.Time) error {
	rows, err := a.db.Query(`SELECT StartTime, EndTime, COALESCE(DurationHours, 0)
		FROM Bookings
		WHERE (PcId = ? OR (PcId IS NULL AND ComputerName = ?)) AND Status IN ('active', 'booked', 'pending')`, pcID, pcName)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var startRaw string
		var endRaw sql.NullString
		var durationHours int
		if err := rows.Scan(&startRaw, &endRaw, &durationHours); err != nil {
			return err
		}
		existingStart, err := time.Parse(time.RFC3339Nano, startRaw)
		if err != nil {
			return err
		}
		existingEnd := existingStart.Add(time.Duration(durationHours) * time.Hour)
		if endRaw.Valid {
			parsedEnd, err := time.Parse(time.RFC3339Nano, endRaw.String)
			if err != nil {
				return err
			}
			existingEnd = parsedEnd
		}
		if startTime.Before(existingEnd) && endTime.After(existingStart) {
			return errBadRequest("pc is already booked for this time")
		}
	}
	return rows.Err()
}

func validateBookingRequest(request BookingRequest) error {
	if request.ClientID <= 0 {
		return errBadRequest("clientId is required")
	}
	if err := requireText(request.PCName, "pcName"); err != nil {
		return err
	}
	if err := requireText(request.StartTime, "startTime"); err != nil {
		return err
	}
	if request.Duration <= 0 {
		return errBadRequest("duration must be greater than zero")
	}
	return nil
}

func validateOrderRequest(request OrderRequest) error {
	if request.ClientID <= 0 {
		return errBadRequest("clientId is required")
	}
	if len(request.Items) == 0 {
		return errBadRequest("items are required")
	}
	for _, item := range request.Items {
		if item.ProductID <= 0 {
			return errBadRequest("productId is required")
		}
		if item.Quantity <= 0 {
			return errBadRequest("quantity must be greater than zero")
		}
	}
	return nil
}

type mobileBookingScanner interface {
	Scan(dest ...any) error
}

func scanMobileBooking(scanner mobileBookingScanner) (Booking, error) {
	var booking Booking
	var startRaw string
	var endRaw sql.NullString
	var hourlyRate float64
	if err := scanner.Scan(&booking.ID, &booking.ClientID, &booking.PCName, &startRaw, &endRaw, &booking.DurationHours, &booking.TotalPrice, &hourlyRate, &booking.Status); err != nil {
		return Booking{}, err
	}

	startTime, err := time.Parse(time.RFC3339Nano, startRaw)
	if err != nil {
		return Booking{}, err
	}
	booking.StartTime = startTime

	if booking.DurationHours <= 0 && endRaw.Valid {
		endTime, err := time.Parse(time.RFC3339Nano, endRaw.String)
		if err != nil {
			return Booking{}, err
		}
		booking.DurationHours = int(endTime.Sub(startTime).Hours())
	}
	if booking.TotalPrice <= 0 && booking.DurationHours > 0 && hourlyRate > 0 {
		booking.TotalPrice = float64(booking.DurationHours) * hourlyRate
	}
	return booking, nil
}
