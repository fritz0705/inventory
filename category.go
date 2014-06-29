package inventory

import (
	"net/http"
	"path"
	"strconv"
)

func (app *Application) ListCategoriesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		app.CreateCategoryHandler(w, r)
		return
	}

	categories, err := LoadCategories(app.Database, `SELECT * FROM 'category'`)
	if err != nil {
		app.Error(w, err)
		return
	}

	app.renderTemplate(w, r, categories, "ListCategories", "Layout")
}

func (app *Application) NewCategoryHandler(w http.ResponseWriter, r *http.Request) {
	app.renderTemplate(w, r, nil, "NewCategory", "Layout")
}

func (app *Application) CreateCategoryHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.Error(w, err)
		return
	}

	category := new(Category)
	category.LoadForm(r.PostForm)

	err = category.Save(app.Database)
	if err != nil {
		app.Error(w, err)
		return
	}

	http.Redirect(w, r, "/categories", http.StatusFound)
}

func (app *Application) EditCategoryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		app.UpdateCategoryHandler(w, r)
		return
	}

	id, _ := strconv.Atoi(path.Base(r.URL.Path))
	if id == 0 {
		app.NotFoundHandler(w, r)
		return
	}

	category, err := FindCategory(app.Database, int64(id))
	if err != nil {
		app.Error(w, err)
		return
	} else if category == nil {
		app.NotFoundHandler(w, r)
		return
	}

	app.renderTemplate(w, r, category, "EditCategory", "Layout")
}

func (app *Application) UpdateCategoryHandler(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(path.Base(r.URL.Path))
	if id == 0 {
		app.NotFoundHandler(w, r)
		return
	}

	err := r.ParseForm()
	if err != nil {
		app.Error(w, err)
		return
	}

	category, err := FindCategory(app.Database, int64(id))
	if err != nil {
		app.Error(w, err)
		return
	} else if category == nil {
		app.NotFoundHandler(w, r)
		return
	}

	err = category.LoadForm(r.PostForm)
	if err != nil {
		app.Error(w, err)
		return
	}

	err = category.Save(app.Database)
	if err != nil {
		app.Error(w, err)
		return
	}

	http.Redirect(w, r, "/categories", http.StatusFound)
}

func (app *Application) DeleteCategoryHandler(w http.ResponseWriter, r *http.Request) {
	_, idStr := path.Split(r.URL.Path)
	id, err := strconv.Atoi(idStr)
	if err != nil {
		app.NotFoundHandler(w, r)
		return
	}

	res, err := app.Database.Exec(`DELETE FROM 'category' WHERE "id" = ?`, id)
	if err != nil {
		app.Error(w, err)
		return
	}

	aff, err := res.RowsAffected()
	if err != nil {
		app.Error(w, err)
		return
	} else if aff == 0 {
		app.NotFoundHandler(w, r)
		return
	}

	http.Redirect(w, r, "/categories", http.StatusSeeOther)
}
