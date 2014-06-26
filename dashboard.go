package inventory

import (
	"net/http"
)

func (app *Application) newestParts() ([]*Part, error) {
	rows, err := app.Database.Query(`SELECT * FROM 'part' ORDER BY 'id' DESC LIMIT 5`)
	if err != nil {
		return nil, err
	}

	parts := make([]*Part, 0)
	for rows.Next() {
		part := new(Part)
		err := part.Load(rows)
		if err != nil {
			return nil, err
		}
		parts = append(parts, part)
	}

	return parts, nil
}

func (app *Application) outOfStockParts() ([]*viewPart, error) {
	rows, err := app.Database.Query(`SELECT 'part'.* FROM 'part_amount' JOIN
	'part' ON 'part'.'id' = 'part_amount'.'part_id' GROUP BY "part_id" HAVING
	"amount" = 0 ORDER BY "timestamp" DESC LIMIT 5`)

	if err != nil {
		return nil, err
	}

	viewParts := make([]*viewPart, 0)
	for rows.Next() {
		part := new(Part)
		err := part.Load(rows)
		if err != nil {
			return nil, err
		}
		viewPart, err := loadViewPart(part, app.Database)
		viewParts = append(viewParts, viewPart)
	}

	return viewParts, nil
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
		"Parts": parts,
		"OutOfStock": outOfStock,
	}, "Dashboard", "Layout")
}
