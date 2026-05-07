package app

import "net/http"

func (a *App) listComputers(w http.ResponseWriter, _ *http.Request) {
	rows, err := a.db.Query(`SELECT Id, Name, IsOccupied FROM Computers ORDER BY Id`)
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

func (a *App) createComputer(w http.ResponseWriter, r *http.Request) {
	var computer Computer
	if err := decodeJSON(r, &computer); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if err := requireText(computer.Name, "name"); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	result, err := a.db.Exec(`INSERT INTO Computers (Name, IsOccupied) VALUES (?, ?)`, computer.Name, computer.IsOccupied)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	computer.ID, _ = result.LastInsertId()
	computer.Status = computerStatus(computer.IsOccupied)
	writeJSON(w, http.StatusCreated, computer)
}

func (a *App) updateComputer(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	var computer Computer
	if err := decodeJSON(r, &computer); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if err := requireText(computer.Name, "name"); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	result, err := a.db.Exec(`UPDATE Computers SET Name = ?, IsOccupied = ? WHERE Id = ?`, computer.Name, computer.IsOccupied, id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		writeError(w, http.StatusNotFound, "computer not found")
		return
	}

	computer.ID = id
	computer.Status = computerStatus(computer.IsOccupied)
	writeJSON(w, http.StatusOK, computer)
}

func (a *App) deleteComputer(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	result, err := a.db.Exec(`DELETE FROM Computers WHERE Id = ?`, id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		writeError(w, http.StatusNotFound, "computer not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func computerStatus(isOccupied bool) string {
	if isOccupied {
		return "Занят"
	}
	return "Свободен"
}
