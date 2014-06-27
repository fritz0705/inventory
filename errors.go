package inventory

import (
	"net/http"
	"log"
)

func (app *Application) NotFoundHandler(w http.ResponseWriter, r *http.Request) {
	// TODO(fritz): Implement pretty NotFound handler
	http.NotFound(w, r)
}

func (app *Application) Error(w http.ResponseWriter, err error) {
	log.Fatal(err)
	http.Error(w, err.Error(), http.StatusInternalServerError)
}
