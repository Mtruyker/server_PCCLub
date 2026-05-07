package app

import (
	"database/sql"
	"net/http"
	"time"
)

func (a *App) listSessions(w http.ResponseWriter, r *http.Request) {
	query := `SELECT Id, ClientName, ComputerName, StartTime, EndTime, TotalCost, COALESCE(TariffName, ''), HourlyRate
		FROM Sessions`
	args := make([]any, 0)

	if computer := r.URL.Query().Get("computer"); computer != "" {
		query += ` WHERE ComputerName = ?`
		args = append(args, computer)
	}
	query += ` ORDER BY Id DESC`

	rows, err := a.db.Query(query, args...)
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

func (a *App) startSession(w http.ResponseWriter, r *http.Request) {
	var request StartSessionRequest
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if err := validateStartSession(request); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	var activeCount int
	err := a.db.QueryRow(`SELECT COUNT(*) FROM Sessions WHERE ComputerName = ? AND EndTime IS NULL`, request.ComputerName).Scan(&activeCount)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if activeCount > 0 {
		writeError(w, http.StatusConflict, "computer already has an active session")
		return
	}

	startTime := time.Now().UTC()
	result, err := a.db.Exec(
		`INSERT INTO Sessions (ClientName, ComputerName, StartTime, EndTime, TotalCost, TariffName, HourlyRate)
		 VALUES (?, ?, ?, NULL, NULL, ?, ?)`,
		request.ClientName, request.ComputerName, startTime.Format(time.RFC3339Nano), request.TariffName, request.HourlyRate,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	_, _ = a.db.Exec(`UPDATE Computers SET IsOccupied = 1 WHERE Name = ?`, request.ComputerName)

	id, _ := result.LastInsertId()
	rate := request.HourlyRate
	session := Session{
		ID:           id,
		ClientName:   request.ClientName,
		ComputerName: request.ComputerName,
		StartTime:    startTime,
		TariffName:   request.TariffName,
		HourlyRate:   &rate,
		Status:       "Активна",
	}

	writeJSON(w, http.StatusCreated, session)
}

func (a *App) completeSession(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	session, err := a.findSession(id)
	if err != nil {
		if err == sql.ErrNoRows {
			writeError(w, http.StatusNotFound, "session not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if session.EndTime != nil {
		writeError(w, http.StatusConflict, "session already completed")
		return
	}

	endTime := time.Now().UTC()
	rate := 50.0
	if session.HourlyRate != nil {
		rate = *session.HourlyRate
	}
	total := endTime.Sub(session.StartTime).Hours() * rate

	_, err = a.db.Exec(`UPDATE Sessions SET EndTime = ?, TotalCost = ? WHERE Id = ?`, endTime.Format(time.RFC3339Nano), total, id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	_, _ = a.db.Exec(`UPDATE Computers SET IsOccupied = 0 WHERE Name = ?`, session.ComputerName)

	session.EndTime = &endTime
	session.TotalCost = &total
	session.Status = "Завершена"
	writeJSON(w, http.StatusOK, session)
}

func (a *App) deleteSession(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	session, err := a.findSession(id)
	if err != nil && err != sql.ErrNoRows {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	result, err := a.db.Exec(`DELETE FROM Sessions WHERE Id = ?`, id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		writeError(w, http.StatusNotFound, "session not found")
		return
	}

	if err == nil && session.EndTime == nil {
		_, _ = a.db.Exec(`UPDATE Computers SET IsOccupied = 0 WHERE Name = ?`, session.ComputerName)
	}

	w.WriteHeader(http.StatusNoContent)
}

func (a *App) findSession(id int64) (Session, error) {
	row := a.db.QueryRow(`SELECT Id, ClientName, ComputerName, StartTime, EndTime, TotalCost, COALESCE(TariffName, ''), HourlyRate FROM Sessions WHERE Id = ?`, id)
	return scanSession(row)
}

func validateStartSession(request StartSessionRequest) error {
	if err := requireText(request.ClientName, "clientName"); err != nil {
		return err
	}
	if err := requireText(request.ComputerName, "computerName"); err != nil {
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

type sessionScanner interface {
	Scan(dest ...any) error
}

func scanSession(scanner sessionScanner) (Session, error) {
	var session Session
	var startRaw string
	var endRaw sql.NullString
	var total sql.NullFloat64
	var rate sql.NullFloat64

	err := scanner.Scan(
		&session.ID,
		&session.ClientName,
		&session.ComputerName,
		&startRaw,
		&endRaw,
		&total,
		&session.TariffName,
		&rate,
	)
	if err != nil {
		return Session{}, err
	}

	startTime, err := time.Parse(time.RFC3339Nano, startRaw)
	if err != nil {
		return Session{}, err
	}
	session.StartTime = startTime

	if endRaw.Valid {
		endTime, err := time.Parse(time.RFC3339Nano, endRaw.String)
		if err != nil {
			return Session{}, err
		}
		session.EndTime = &endTime
		session.Status = "Завершена"
	} else {
		session.Status = "Активна"
	}

	if total.Valid {
		session.TotalCost = &total.Float64
	}
	if rate.Valid {
		session.HourlyRate = &rate.Float64
	}

	return session, nil
}
