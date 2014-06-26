package inventory

import (
	"net/http"
)

func (app *Application) SearchHandler(w http.ResponseWriter, r *http.Request) {
	query := r.FormValue("query")
	_ = query
}
