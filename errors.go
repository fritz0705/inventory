package inventory

import (
	"database/sql"
	"log"
	"net/http"
)

func (app *Application) NotFoundHandler(w http.ResponseWriter, r *http.Request) {
	// TODO(fritz): Implement pretty NotFound handler
	http.NotFound(w, r)
}

func (app *Application) Error(w http.ResponseWriter, err error) {
	log.Print(err)
	http.Error(w, err.Error(), http.StatusInternalServerError)
}

func (app *Application) SQLError(w http.ResponseWriter, r *http.Request, err error) bool {
	switch err {
	case sql.ErrNoRows:
		app.NotFoundHandler(w, r)
		return true
	case nil:
	default:
		app.Error(w, err)
		return true
	}

	return false
}
