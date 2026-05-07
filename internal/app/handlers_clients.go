package app

import "net/http"

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

func (a *App) createClient(w http.ResponseWriter, r *http.Request) {
	var client Client
	if err := decodeJSON(r, &client); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if err := requireText(client.Name, "name"); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	result, err := a.db.Exec(
		`INSERT INTO Clients (Name, Phone, Email, Balance) VALUES (?, ?, ?, ?)`,
		client.Name, client.Phone, client.Email, client.Balance,
	)
	if err != nil {
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

	var client Client
	if err := decodeJSON(r, &client); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if err := requireText(client.Name, "name"); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	result, err := a.db.Exec(
		`UPDATE Clients SET Name = ?, Phone = ?, Email = ?, Balance = ? WHERE Id = ?`,
		client.Name, client.Phone, client.Email, client.Balance, id,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		writeError(w, http.StatusNotFound, "client not found")
		return
	}

	client.ID = id
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
