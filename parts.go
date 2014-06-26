package inventory

import (
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"
)

type viewPart struct {
	*Part
	Category     *Category
	LatestAmount *PartAmount
	Place        *Place
}

func loadViewPart(part *Part, db Queryer) (*viewPart, error) {
	var err error

	viewPart := new(viewPart)
	viewPart.Part = part
	viewPart.Category, err = part.Category(db)
	viewPart.Place, err = part.Place(db)
	if err != nil {
		return viewPart, err
	}

	viewPart.LatestAmount, err = part.LatestAmount(db)
	return viewPart, err
}

func buildPartsQuery(form url.Values) (string, []interface{}) {
	query := []string{}
	args := make([]interface{}, 0)
	if form["category"] != nil {
		subQuery := []string{}
		for _, category := range form["category"] {
			subQuery = append(subQuery, `"category_id" = ?`)
			args = append(args, category)
		}

		query = append(query, "("+strings.Join(subQuery, " OR ")+")")
	}

	if len(query) == 0 {
		query = append(query, "1 = 1")
	}

	return strings.Join(query, " AND "), args
}

func buildPartAmountGraph(amounts []*PartAmount) (res [][2]int64) {
	for _, amount := range amounts {
		res = append(res, [2]int64{amount.Timestamp.Unix() * 1000, amount.Amount})
	}
	return res
}

func (app *Application) ListPartsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		app.CreatePartHandler(w, r)
		return
	}

	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	filterQuery, filterArgs := buildPartsQuery(r.Form)

	rows, err := app.Database.Query(`SELECT * FROM 'part' WHERE (`+filterQuery+`) ORDER BY "id" DESC`, filterArgs...)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	parts, err := LoadParts(rows)
	viewParts := make([]*viewPart, len(parts))
	for n, part := range parts {
		viewParts[n], err = loadViewPart(part, app.Database)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	categories, err := LoadCategories(app.Database, "SELECT * FROM 'category'")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	app.renderTemplate(w, r, map[string]interface{}{
		"Parts":      viewParts,
		"Categories": categories,
	}, "ListParts", "Layout")
}

func (app *Application) NewPartHandler(w http.ResponseWriter, r *http.Request) {
	categories, err := LoadCategories(app.Database, "SELECT * FROM 'category'")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	places, err := LoadPlaces(app.Database, "SELECT * FROM 'place'")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	app.renderTemplate(w, r, map[string]interface{}{
		"Obj":        &Part{},
		"Categories": categories,
		"Places":     places,
	}, "NewPart", "Layout")
}

func (app *Application) ShowPartHandler(w http.ResponseWriter, r *http.Request) {
	_, idString := path.Split(r.URL.Path)
	id, err := strconv.Atoi(idString)
	if err != nil {
		app.NotFoundHandler(w, r)
		return
	}
	app.renderTemplate(w, r, id, "ShowPart", "Layout")
}

func (app *Application) EditPartHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		app.UpdatePartHandler(w, r)
		return
	}

	id, err := strconv.Atoi(path.Base(r.URL.Path))
	if err != nil {
		app.NotFoundHandler(w, r)
		return
	}

	rows, err := app.Database.Query(`SELECT * FROM 'part' WHERE "id" = ? LIMIT 1`, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	if !rows.Next() {
		app.NotFoundHandler(w, r)
		return
	}

	part := new(Part)
	err = part.Load(rows)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	places, err := LoadPlaces(app.Database, "SELECT * FROM 'place'")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	categories, err := LoadCategories(app.Database, "SELECT * FROM 'category'")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	partAmounts, err := LoadPartAmounts(app.Database, `SELECT * FROM 'part_amount'
	WHERE "part_id" = ? ORDER BY 'timestamp' DESC LIMIT 30`, part.Id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	app.renderTemplate(w, r, map[string]interface{}{
		"Obj":         part,
		"Categories":  categories,
		"Places":      places,
		"AmountGraph": buildPartAmountGraph(partAmounts),
	}, "EditPart", "Layout")
}

func (app *Application) CreatePartAmountHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	partId, err := strconv.Atoi(path.Base(r.URL.Path))
	if err != nil {
		app.NotFoundHandler(w, r)
		return
	}

	amount, err := strconv.Atoi(r.PostFormValue("amount"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tx, err := app.Database.Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Commit()

	part, err := FindPart(tx, int64(partId))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	} else if part == nil {
		app.NotFoundHandler(w, r)
		return
	}

	partAmount := &PartAmount{
		PartId:    part.Id,
		Amount:    int64(amount),
		Timestamp: time.Now(),
	}

	err = partAmount.Save(tx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tx.Commit()

	http.Redirect(w, r, fmt.Sprintf("/parts/edit/%d", part.Id), http.StatusSeeOther)
}

func (app *Application) UpdatePartHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(path.Base(r.URL.Path))
	if err != nil {
		app.NotFoundHandler(w, r)
		return
	}

	err = r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	part, err := FindPart(app.Database, int64(id))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	} else if part == nil {
		app.NotFoundHandler(w, r)
		return
	}

	err = part.LoadForm(r.PostForm)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = part.Save(app.Database)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/parts/edit/%d", id), http.StatusFound)
}

func (app *Application) CreatePartHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tx, err := app.Database.Begin()

	// Create Part object
	part := new(Part)
	err = part.LoadForm(r.PostForm)

	err = part.Save(tx)
	if err != nil {
		tx.Rollback()
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	amount, err := strconv.Atoi(r.PostForm.Get("amount"))
	if err != nil {
		amount = 0
	}

	// Create PartAmount object
	partAmount := &PartAmount{
		PartId:    part.Id,
		Amount:    int64(amount),
		Timestamp: time.Now(),
	}

	err = partAmount.Save(tx)
	if err != nil {
		tx.Rollback()
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tx.Commit()

	http.Redirect(w, r, "/parts", http.StatusFound)
}

func (app *Application) EmptyPartHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(path.Base(r.URL.Path))
	if err != nil {
		app.NotFoundHandler(w, r)
		return
	}

	rows, err := app.Database.Query(`SELECT * FROM 'part' WHERE "id" = ? LIMIT 1`, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !rows.Next() {
		app.NotFoundHandler(w, r)
		return
	}
	rows.Close()

	partAmount := &PartAmount{
		PartId:    int64(id),
		Amount:    0,
		Timestamp: time.Now(),
	}

	err = partAmount.Save(app.Database)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/parts", http.StatusSeeOther)
}

func (app *Application) DeletePartHandler(w http.ResponseWriter, r *http.Request) {
	_, idStr := path.Split(r.URL.Path)
	id, err := strconv.Atoi(idStr)
	if err != nil {
		app.NotFoundHandler(w, r)
		return
	}

	res, err := app.Database.Exec(`DELETE FROM 'part' WHERE "id" = ?`, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	aff, err := res.RowsAffected()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if aff == 0 {
		app.NotFoundHandler(w, r)
		return
	}

	http.Redirect(w, r, "/parts", http.StatusSeeOther)
}
