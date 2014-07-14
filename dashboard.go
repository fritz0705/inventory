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

	app.renderTemplate(w, r, map[string]interface{}{
		"Parts":      parts,
		"OutOfStock": outOfStock,
	}, "Dashboard", "Layout")
}
