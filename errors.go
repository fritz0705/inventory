package inventory

import (
	"net/http"
)

func (app *Application) NotFoundHandler(w http.ResponseWriter, r *http.Request) {
	// TODO(fritz): Implement pretty NotFound handler
	http.NotFound(w, r)
}
