package inventory

import (
	"database/sql"
	"net/http"
	"path"
	"strconv"
)

func (app *Application) ListPlacesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		app.CreatePlaceHandler(w, r)
		return
	}

	places := []Place{}
	err := app.DB.Select(&places, `SELECT * FROM 'place'`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	app.renderTemplate(w, r, places, "ListPlaces", "Layout")
}

func (app *Application) CreatePlaceHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	place := new(Place)
	err = place.LoadForm(r.PostForm)

	err = place.Save(app.DB)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/places", http.StatusFound)
}

func (app *Application) NewPlaceHandler(w http.ResponseWriter, r *http.Request) {
	app.renderTemplate(w, r, nil, "NewPlace", "Layout")
}

func (app *Application) UpdatePlaceHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		panic(err)
	}

	tx := app.DB.MustBegin()
	defer tx.Rollback()

	id := path.Base(r.URL.Path)
	if id == "" {
		app.NotFoundHandler(w, r)
		return
	}

	place := new(Place)
	err = tx.Get(place, `SELECT * FROM 'place' WHERE "id" = ?`, id)
	switch err {
	case nil:
	case sql.ErrNoRows:
		app.NotFoundHandler(w, r)
		return
	default:
		app.Error(w, err)
		return
	}

	place.LoadForm(r.PostForm)

	err = place.Save(tx)
	if err != nil {
		app.Error(w, err)
	}

	tx.Commit()

	http.Redirect(w, r, "/places", http.StatusSeeOther)
}

func (app *Application) EditPlaceHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		app.UpdatePlaceHandler(w, r)
		return
	}

	id := path.Base(r.URL.Path)
	if id == "" {
		app.NotFoundHandler(w, r)
		return
	}

	place := new(Place)
	err := app.DB.Get(place, `SELECT * FROM 'place' WHERE "id" = ?`, id)
	switch err {
	case sql.ErrNoRows:
		app.NotFoundHandler(w, r)
		return
	case nil:
	default:
		app.Error(w, err)
		return
	}

	app.renderTemplate(w, r, map[string]interface{}{
		"Place": place,
	}, "EditPlace", "Layout")
}

func (app *Application) DeletePlaceHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		app.NotFoundHandler(w, r)
		return
	}

	_, idStr := path.Split(r.URL.Path)
	id, err := strconv.Atoi(idStr)
	if err != nil {
		app.NotFoundHandler(w, r)
		return
	}

	res, err := app.DB.Exec(`DELETE FROM 'place' WHERE "id" = ?`, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	aff, err := res.RowsAffected()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	} else if aff == 0 {
		app.NotFoundHandler(w, r)
		return
	}

	http.Redirect(w, r, "/places", http.StatusSeeOther)
}
