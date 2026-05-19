package app

import (
	"database/sql"
	"net/http"
	"strings"
)

type ClientProfileRequest struct {
	Name  string `json:"name"`
	Phone string `json:"phone"`
	Email string `json:"email"`
}

type TopUpRequest struct {
	Amount float64 `json:"amount"`
}

func (a *App) listClients(w http.ResponseWriter, _ *http.Request) {
	rows, err := a.db.Query(`SELECT Id, Name, COALESCE(Phone, ''), COALESCE(Email, ''), Balance FROM Clients ORDER BY Id`)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	clients := make([]Client, 0)
	for rows.Next() {
		var client Client
		if err := rows.Scan(&client.ID, &client.Name, &client.Phone, &client.Email, &client.Balance); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		clients = append(clients, client)
	}

	writeJSON(w, http.StatusOK, clients)
}

func (a *App) getClient(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := a.requireClient(r, id); err != nil {
		writeAuthError(w, err)
		return
	}

	client, err := a.findClientByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			writeError(w, http.StatusNotFound, "client not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, client)
}

func (a *App) createClient(w http.ResponseWriter, r *http.Request) {
	var client Client
	if err := decodeJSON(r, &client); err != nil {
		writeErrorCode(w, http.StatusBadRequest, "INVALID_JSON", "Некорректный JSON")
		return
	}
	client.Name = strings.TrimSpace(client.Name)
	client.Phone = normalizePhone(client.Phone)
	client.Email = strings.TrimSpace(client.Email)
	if err := validateClientFields(client.Name, client.Phone, client.Email, false); err != nil {
		writeValidationError(w, err)
		return
	}

	result, err := a.db.Exec(
		`INSERT INTO Clients (Name, Phone, Email, Balance) VALUES (?, ?, ?, ?)`,
		client.Name, client.Phone, client.Email, client.Balance,
	)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "unique") {
			writeErrorCode(w, http.StatusConflict, "CLIENT_EXISTS", "Телефон или email уже зарегистрирован")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	client.ID, _ = result.LastInsertId()
	writeJSON(w, http.StatusCreated, client)
}

func (a *App) updateClient(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	isAdmin, err := a.requireClientOrAdmin(r, id)
	if err != nil {
		writeAuthError(w, err)
		return
	}

	if isAdmin {
		a.updateClientAsAdmin(w, r, id)
		return
	}

	a.updateClientProfile(w, r, id)
}

func (a *App) updateClientProfile(w http.ResponseWriter, r *http.Request, id int64) {
	var request ClientProfileRequest
	if err := decodeJSON(r, &request); err != nil {
		writeErrorCode(w, http.StatusBadRequest, "INVALID_JSON", "Некорректный JSON")
		return
	}
	request.Name = strings.TrimSpace(request.Name)
	request.Phone = normalizePhone(request.Phone)
	request.Email = strings.TrimSpace(request.Email)
	if err := validateClientFields(request.Name, request.Phone, request.Email, false); err != nil {
		writeValidationError(w, err)
		return
	}

	result, err := a.db.Exec(
		`UPDATE Clients SET Name = ?, Phone = ?, Email = ? WHERE Id = ?`,
		request.Name, request.Phone, request.Email, id,
	)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "unique") {
			writeErrorCode(w, http.StatusConflict, "CLIENT_EXISTS", "Телефон или email уже зарегистрирован")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		writeError(w, http.StatusNotFound, "client not found")
		return
	}

	client, err := a.findClientByID(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, client)
}

func (a *App) updateClientAsAdmin(w http.ResponseWriter, r *http.Request, id int64) {
	var client Client
	if err := decodeJSON(r, &client); err != nil {
		writeErrorCode(w, http.StatusBadRequest, "INVALID_JSON", "Некорректный JSON")
		return
	}
	client.Name = strings.TrimSpace(client.Name)
	client.Phone = normalizePhone(client.Phone)
	client.Email = strings.TrimSpace(client.Email)
	if err := validateClientFields(client.Name, client.Phone, client.Email, false); err != nil {
		writeValidationError(w, err)
		return
	}

	result, err := a.db.Exec(
		`UPDATE Clients SET Name = ?, Phone = ?, Email = ?, Balance = ? WHERE Id = ?`,
		client.Name, client.Phone, client.Email, client.Balance, id,
	)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "unique") {
			writeErrorCode(w, http.StatusConflict, "CLIENT_EXISTS", "Телефон или email уже зарегистрирован")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		writeError(w, http.StatusNotFound, "client not found")
		return
	}

	client, err = a.findClientByID(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, client)
}

func (a *App) topUpClient(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if _, err := a.requireClientOrAdmin(r, id); err != nil {
		writeAuthError(w, err)
		return
	}

	var request TopUpRequest
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if request.Amount <= 0 {
		writeError(w, http.StatusBadRequest, "amount must be greater than zero")
		return
	}

	result, err := a.db.Exec(`UPDATE Clients SET Balance = Balance + ? WHERE Id = ?`, request.Amount, id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		writeError(w, http.StatusNotFound, "client not found")
		return
	}

	client, err := a.findClientByID(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, client)
}

func (a *App) deleteClient(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	result, err := a.db.Exec(`DELETE FROM Clients WHERE Id = ?`, id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		writeError(w, http.StatusNotFound, "client not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (a *App) findClientByID(id int64) (Client, error) {
	var client Client
	err := a.db.QueryRow(`SELECT Id, Name, COALESCE(Phone, ''), COALESCE(Email, ''), Balance FROM Clients WHERE Id = ?`, id).
		Scan(&client.ID, &client.Name, &client.Phone, &client.Email, &client.Balance)
	return client, err
}
