package app

import "net/http"

func (a *App) listTariffs(w http.ResponseWriter, _ *http.Request) {
	rows, err := a.db.Query(`SELECT Id, Name, CostPerHour, COALESCE(Description, '') FROM Tariffs ORDER BY Id`)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	tariffs := make([]Tariff, 0)
	for rows.Next() {
		var tariff Tariff
		if err := rows.Scan(&tariff.ID, &tariff.Name, &tariff.CostPerHour, &tariff.Description); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		tariffs = append(tariffs, tariff)
	}

	writeJSON(w, http.StatusOK, tariffs)
}

func (a *App) createTariff(w http.ResponseWriter, r *http.Request) {
	var tariff Tariff
	if err := decodeJSON(r, &tariff); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if err := validateTariff(tariff); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	result, err := a.db.Exec(`INSERT INTO Tariffs (Name, CostPerHour, Description) VALUES (?, ?, ?)`, tariff.Name, tariff.CostPerHour, tariff.Description)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	tariff.ID, _ = result.LastInsertId()
	writeJSON(w, http.StatusCreated, tariff)
}

func (a *App) updateTariff(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	var tariff Tariff
	if err := decodeJSON(r, &tariff); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if err := validateTariff(tariff); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	result, err := a.db.Exec(`UPDATE Tariffs SET Name = ?, CostPerHour = ?, Description = ? WHERE Id = ?`, tariff.Name, tariff.CostPerHour, tariff.Description, id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		writeError(w, http.StatusNotFound, "tariff not found")
		return
	}

	tariff.ID = id
	writeJSON(w, http.StatusOK, tariff)
}

func (a *App) deleteTariff(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	result, err := a.db.Exec(`DELETE FROM Tariffs WHERE Id = ?`, id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		writeError(w, http.StatusNotFound, "tariff not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func validateTariff(tariff Tariff) error {
	if err := requireText(tariff.Name, "name"); err != nil {
		return err
	}
	if tariff.CostPerHour <= 0 {
		return errBadRequest("costPerHour must be greater than zero")
	}
	return nil
}

type errBadRequest string

func (e errBadRequest) Error() string {
	return string(e)
}
