package inventory

import (
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
