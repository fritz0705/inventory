package inventory

import (
	"net/http"
)

func (app *Application) newestParts() ([]PartView, error) {
	var partViews []PartView
	err := app.DB.Select(&partViews, `SELECT * FROM 'part_view' ORDER BY "created_at" DESC`)
	return partViews, err
}

func (app *Application) outOfStockParts() ([]PartView, error) {
	var partViews []PartView
	err := app.DB.Select(&partViews, `SELECT * FROM 'part_view' WHERE "amount" = 0`)
	return partViews, err
}

func (app *Application) statisticsPanel() (map[string]interface{}, error) {
	var (
		totalParts int64
		totalStock int64
		emptyParts int64
		totalPlaces int64
		totalCategories int64
	)

	row := app.DB.QueryRowx(`SELECT COUNT(*), SUM(amount) FROM 'part_view'`)
	if err := row.Scan(&totalParts, &totalStock); err != nil {
		return nil, err
	}

	row = app.DB.QueryRowx(`SELECT COUNT(*) FROM 'part_view' WHERE "amount" = 0`)
	if err := row.Scan(&emptyParts); err != nil {
		return nil, err
	}

	row = app.DB.QueryRowx(`SELECT COUNT(*) FROM 'place'`)
	if err := row.Scan(&totalPlaces); err != nil {
		return nil, err
	}

	row = app.DB.QueryRowx(`SELECT COUNT(*) FROM 'category'`)
	if err := row.Scan(&totalCategories); err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"TotalParts": totalParts,
		"TotalStock": totalStock,
		"EmptyParts": emptyParts,
		"TotalPlaces": totalPlaces,
		"TotalCategories": totalCategories,
	}, nil
}

func (app *Application) DashboardHandler(w http.ResponseWriter, r *http.Request) {
	parts, err := app.newestParts()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	outOfStock, err := app.outOfStockParts()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	statistics, err := app.statisticsPanel()
	if err != nil {
		app.Error(w, err)
		return
	}

	app.renderTemplate(w, r, map[string]interface{}{
		"Parts":      parts,
		"OutOfStock": outOfStock,
		"Statistics": statistics,
	}, "Dashboard", "Layout")
}
