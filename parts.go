package inventory

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/fritz0705/inventory/si"
)

var PartsPerPage = 10

func buildPartAmountGraph(amounts []PartAmount) (res [][2]int64) {
	for _, amount := range amounts {
		res = append(res, [2]int64{amount.Timestamp.Unix() * 1000, amount.Amount})
	}
	return res
}

func splitSiRange(r string) (res [2]si.Number, err error) {
	parts := strings.SplitN(r, "-", 2)
	if len(parts) != 2 {
		panic("inventory: Invalid operation splitSiRange on non-range string")
	}
	left, right := parts[0], parts[1]

	leftVal, err := si.Parse(left)
	if err != nil {
		return
	}
	rightVal, err := si.Parse(right)
	if err != nil {
		return
	}

	res = [2]si.Number{leftVal, rightVal}
	return
}

func buildListPartsQuery(r *http.Request) (query string, args []interface{}, err error) {
	err = r.ParseForm()
	if err != nil {
		return
	}

	query += `SELECT * FROM 'part_view' WHERE (1=1)`

	lastId, _ := strconv.Atoi(r.Form.Get("last_id"))
	firstId, _ := strconv.Atoi(r.Form.Get("first_id"))

	if lastId != 0 {
		query += ` AND "id" > ?`
		args = append(args, lastId)
	} else if firstId != 0 {
		query += ` AND "id" < ?`
		args = append(args, firstId)
	}

	if r.Form["category"] != nil {
		var categoryList []string

		for _, category := range r.Form["category"] {
			category, _ := strconv.Atoi(category)
			if category != 0 {
				categoryList = append(categoryList, strconv.Itoa(category))
			}
		}

		query += ` AND "category_id" IN (` + strings.Join(categoryList, ", ") + `)`
	}

	if value := r.Form.Get("value"); value != "" {
		if strings.ContainsRune(value, '-') {
			// Range query
			rng, err := splitSiRange(value)
			if err == nil {
				query += ` AND "value" BETWEEN ? AND ?`
				args = append(args, rng[0].Value(), rng[1].Value())
			}
		} else {
			// Value query
			value, err := si.Parse(value)
			if err == nil {
				query += ` AND "value" = ?`
				args = append(args, value.Value())
			}
		}
	}

	if amount := r.Form.Get("amount"); amount != "" {
		if strings.ContainsRune(amount, '-') {
			// Range query
			rng, err := splitSiRange(amount)
			if err == nil {
				query += ` AND "amount" BETWEEN ? AND ?`
				args = append(args, rng[0].Value(), rng[1].Value())
			}
		} else {
			// Value query
			value, err := si.Parse(amount)
			if err == nil {
				query += ` AND "amount" = ?`
				args = append(args, value.Value())
			}
		}
	}

	query += ` ORDER BY "id" DESC LIMIT ` + strconv.Itoa(PartsPerPage)
	if r.Form["page"] != nil {
		page, _ := strconv.Atoi(r.Form.Get("page"))
		if page != 0 {
			query += ` OFFSET ?`
			args = append(args, page*PartsPerPage)
		}
	}

	return
}

func pageQuerys(url *url.URL, cur int) (string, string) {
	values := url.Query()
	values.Set("page", strconv.Itoa(cur-1))
	prev := values.Encode()
	values.Set("page", strconv.Itoa(cur+1))
	return prev, values.Encode()
}

func (app *Application) ListPartsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		app.CreatePartHandler(w, r)
		return
	}

	query, args, err := buildListPartsQuery(r)
	if err != nil {
		app.Error(w, err)
		return
	}

	partViews := []PartView{}
	err = app.DB.Select(&partViews, query, args...)
	if err != nil {
		app.Error(w, err)
		return
	}

	categories := []Category{}
	err = app.DB.Select(&categories, `SELECT * FROM 'category'`)
	if err != nil {
		app.Error(w, err)
		return
	}

	currentPage, _ := strconv.Atoi(r.FormValue("page"))
	prevQuery, nextQuery := pageQuerys(r.URL, currentPage)

	app.renderTemplate(w, r, map[string]interface{}{
		"Parts":       partViews,
		"Categories":  categories,
		"CurrentPage": currentPage,
		"NextPage":    template.URL(nextQuery),
		"PrevPage":    template.URL(prevQuery),
		"URL":         r.URL,
	}, "ListParts", "Layout")
}

func (app *Application) NewPartHandler(w http.ResponseWriter, r *http.Request) {
	categories := []Category{}
	err := app.DB.Select(&categories, `SELECT * FROM 'category'`)
	if err != nil {
		app.Error(w, err)
		return
	}

	places := []Place{}
	err = app.DB.Select(&places, `SELECT * FROM 'place'`)
	if err != nil {
		app.Error(w, err)
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

	var err error

	id := path.Base(r.URL.Path)

	partView := new(PartView)
	err = app.DB.Get(partView, `SELECT * FROM 'part_view' WHERE "id" = ?`, id)
	switch err {
	case sql.ErrNoRows:
		app.NotFoundHandler(w, r)
		return
	case nil:
	default:
		app.Error(w, err)
		return
	}

	places := []Place{}
	err = app.DB.Select(&places, `SELECT * FROM 'place'`)
	if err != nil {
		app.Error(w, err)
		return
	}

	categories := []Category{}
	err = app.DB.Select(&categories, `SELECT * FROM 'category'`)
	if err != nil {
		app.Error(w, err)
		return
	}

	partAmounts := []PartAmount{}
	err = app.DB.Select(&partAmounts, `SELECT * FROM 'part_amount'
	WHERE "part_id" = ? ORDER BY "timestamp" DESC LIMIT 30`, partView.Id)
	if err != nil {
		app.Error(w, err)
		return
	}

	attachments := []Attachment{}
	err = app.DB.Select(&attachments, `SELECT * FROM 'attachment'
	WHERE "part_id" = ? ORDER BY "created_at" ASC`, partView.Id)
	if err != nil {
		app.Error(w, err)
		return
	}

	app.renderTemplate(w, r, map[string]interface{}{
		"Part":        partView,
		"Categories":  categories,
		"Places":      places,
		"AmountGraph": buildPartAmountGraph(partAmounts),
		"Attachments": attachments,
	}, "EditPart", "Layout")
}

func (app *Application) CreatePartAmountHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.Error(w, err)
		return
	}

	partId, err := strconv.Atoi(path.Base(r.URL.Path))
	if err != nil {
		app.NotFoundHandler(w, r)
		return
	}

	amount, err := strconv.Atoi(r.PostFormValue("amount"))
	if err != nil {
		app.Error(w, err)
		return
	}

	part := new(Part)
	err = app.DB.Get(part, `SELECT * FROM 'part' WHERE "id" = ?`, partId)
	switch err {
	case sql.ErrNoRows:
		app.NotFoundHandler(w, r)
		return
	case nil:
	default:
		app.Error(w, err)
		return
	}

	partAmount := &PartAmount{
		PartId:    part.Id,
		Amount:    int64(amount),
		Timestamp: time.Now(),
	}

	err = partAmount.Save(app.DB)
	if err != nil {
		app.Error(w, err)
		return
	}

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
		app.Error(w, err)
		return
	}

	part := new(Part)
	err = app.DB.Get(part, `SELECT * FROM 'part' WHERE "id" = ?`, id)
	switch err {
	case sql.ErrNoRows:
		app.NotFoundHandler(w, r)
		return
	case nil:
	default:
		app.Error(w, err)
		return
	}

	err = part.LoadForm(r.PostForm)
	if err != nil {
		app.Error(w, err)
		return
	}

	err = part.Save(app.DB)
	if err != nil {
		app.Error(w, err)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/parts/edit/%d", id), http.StatusFound)
}

func (app *Application) CreatePartHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.Error(w, err)
		return
	}

	tx, err := app.DB.Begin()

	// Create Part object
	part := new(Part)
	part.CreatedAt = time.Now()
	err = part.LoadForm(r.PostForm)

	err = part.Save(tx)
	if err != nil {
		tx.Rollback()
		app.Error(w, err)
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
		app.Error(w, err)
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

	rows, err := app.DB.Query(`SELECT * FROM 'part' WHERE "id" = ? LIMIT 1`, id)
	if err != nil {
		app.Error(w, err)
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

	err = partAmount.Save(app.DB)
	if err != nil {
		app.Error(w, err)
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

	res, err := app.DB.Exec(`DELETE FROM 'part' WHERE "id" = ?`, id)
	if err != nil {
		app.Error(w, err)
		return
	}

	aff, err := res.RowsAffected()
	if err != nil {
		app.Error(w, err)
		return
	}

	if aff == 0 {
		app.NotFoundHandler(w, r)
		return
	}

	http.Redirect(w, r, "/parts", http.StatusSeeOther)
}
