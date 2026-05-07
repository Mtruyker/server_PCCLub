package app

import "net/http"

func (a *App) statistics(w http.ResponseWriter, _ *http.Request) {
	stats := Statistics{}

	if err := a.db.QueryRow(`SELECT COALESCE(SUM(TotalCost), 0), COUNT(*) FROM Sessions WHERE EndTime IS NOT NULL`).Scan(&stats.TotalRevenue, &stats.CompletedSessions); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if err := a.db.QueryRow(`SELECT COUNT(*) FROM Sessions WHERE EndTime IS NULL`).Scan(&stats.ActiveSessions); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	var err error
	stats.DailyRevenue, err = a.statRows(`SELECT substr(StartTime, 1, 10), COALESCE(SUM(TotalCost), 0)
		FROM Sessions
		WHERE EndTime IS NOT NULL
		GROUP BY substr(StartTime, 1, 10)
		ORDER BY substr(StartTime, 1, 10) DESC
		LIMIT 14`)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	stats.ComputerDistribution, err = a.statRows(`SELECT ComputerName, COUNT(*)
		FROM Sessions
		GROUP BY ComputerName
		ORDER BY COUNT(*) DESC`)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	stats.PopularTariffs, err = a.statRows(`SELECT COALESCE(NULLIF(TariffName, ''), 'Без тарифа'), COUNT(*)
		FROM Sessions
		GROUP BY COALESCE(NULLIF(TariffName, ''), 'Без тарифа')
		ORDER BY COUNT(*) DESC`)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, stats)
}

func (a *App) statRows(query string) ([]StatisticItem, error) {
	rows, err := a.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]StatisticItem, 0)
	for rows.Next() {
		var item StatisticItem
		if err := rows.Scan(&item.Name, &item.Value); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, rows.Err()
}
