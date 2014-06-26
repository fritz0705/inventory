package inventory

import (
	"fmt"
	"net/http"
)

func (app *Application) SearchHandler(w http.ResponseWriter, r *http.Request) {
	query := r.FormValue("query")
	print(query)

	// First Step: Look for part which has the same name or id
	part, err := LoadPart(app.Database, `SELECT * FROM 'part' WHERE "name" = ? LIMIT 1`, query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if part != nil {
		http.Redirect(w, r, fmt.Sprintf("/parts/edit/%d", part.Id), http.StatusFound)
		return
	}

	type SearchQuery struct {
		Unit     string
		Value    float64
		Keywords []string
	}
}
