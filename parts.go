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

type siRange struct {
	Low  si.Number
	High si.Number
}

func parseSiRange(s string) (r *siRange, err error) {
	r = new(siRange)
	if strings.ContainsRune(s, '-') {
		parts := strings.SplitN(s, "-", 2)
		left, right := parts[0], parts[1]

		r.Low, err = si.Parse(left)
		if err != nil {
			return
		}

		r.High, err = si.Parse(right)
	} else {
		r.High, err = si.Parse(s)
		r.Low = r.High
	}

	return
}

func (s siRange) IsEmpty() bool {
	return s.Low == s.High
}

func (s siRange) String() string {
	if s.Low == s.High {
		return s.Low.String()
	}
	return s.Low.String() + "-" + s.High.String()
}

type partsFilter struct {
	Categories map[int64]bool
	Places     map[int64]bool
	Value      *siRange
	Name       string
	Stock      *siRange
}

func loadPartsFilter(form url.Values) (filter *partsFilter, err error) {
	filter = &partsFilter{
		Categories: make(map[int64]bool),
		Places:     make(map[int64]bool),
	}
	for key, value := range form {
		val := value[0]
		if val == "" {
			continue
		}
		switch key {
		case "value":
			filter.Value, err = parseSiRange(val)
		case "amount":
			filter.Stock, err = parseSiRange(val)
		case "name":
			filter.Name = val
		case "category":
			for _, val := range value {
				category, _ := strconv.Atoi(val)
				if category != 0 {
					filter.Categories[int64(category)] = true
				}
			}
		case "place":
			for _, val := range value {
				place, _ := strconv.Atoi(val)
				if place != 0 {
					filter.Places[int64(place)] = true
				}
			}
		}
		if err != nil {
			return
		}
	}
	return
}

func (f partsFilter) CategoriesList() []string {
	res := make([]string, len(f.Categories))
	n := 0
	for id := range f.Categories {
		res[n] = strconv.Itoa(int(id))
		n++
	}
	return res
}

func (f partsFilter) PlacesList() []string {
	res := make([]string, len(f.Places))
	n := 0
	for id := range f.Places {
		res[n] = strconv.Itoa(int(id))
		n++
	}
	return res
}

func buildListPartsQuery(filter *partsFilter, form url.Values) (query string, args []interface{}, err error) {
	query += `SELECT * FROM 'part_view' WHERE (1=1)`

	lastId, _ := strconv.Atoi(form.Get("last_id"))
	firstId, _ := strconv.Atoi(form.Get("first_id"))

	if lastId != 0 {
		query += ` AND "id" > ?`
		args = append(args, lastId)
	} else if firstId != 0 {
		query += ` AND "id" < ?`
		args = append(args, firstId)
	}

	if len(filter.Categories) != 0 {
		query += ` AND "category_id" IN (` + strings.Join(filter.CategoriesList(), ", ") + `)`
	}

	if len(filter.Places) != 0 {
		query += ` AND "place_id" IN (` + strings.Join(filter.PlacesList(), ", ") + `)`
	}

	if filter.Value != nil {
		if filter.Value.IsEmpty() {
			query += ` AND "value" = ?`
			args = append(args, filter.Value.Low.Value())
		} else {
			query += ` AND "value" BETWEEN ? AND ?`
			args = append(args, filter.Value.Low.Value(), filter.Value.High.Value())
		}
	}

	if filter.Stock != nil {
		if filter.Stock.IsEmpty() {
			query += ` AND "amount" = ?`
			args = append(args, filter.Stock.Low.Value())
		} else {
			query += ` AND "value" BETWEEN ? AND ?`
			args = append(args, filter.Stock.Low.Value(), filter.Stock.High.Value())
		}
	}

	if filter.Name != "" {
		query += ` AND "name" GLOB ?`
		args = append(args, filter.Name)
	}

	query += ` ORDER BY "id" DESC LIMIT ` + strconv.Itoa(PartsPerPage)
	if form["page"] != nil {
		page, _ := strconv.Atoi(form.Get("page"))
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

	err := r.ParseForm()
	if err != nil {
		app.Error(w, err)
		return
	}

	filter, err := loadPartsFilter(r.Form)
	if err != nil {
		app.Error(w, err)
		return
	}

	tx := app.DB.MustBegin()
	defer tx.Rollback()

	query, args, err := buildListPartsQuery(filter, r.Form)
	if err != nil {
		app.Error(w, err)
		return
	}

	partViews := []PartView{}
	err = tx.Select(&partViews, query, args...)
	if err != nil {
		app.Error(w, err)
		return
	}

	categories := []Category{}
	err = tx.Select(&categories, `SELECT * FROM 'category'`)
	if err != nil {
		app.Error(w, err)
		return
	}

	places := []Place{}
	err = tx.Select(&places, `SELECT * FROM 'place'`)
	if err != nil {
		app.Error(w, err)
		return
	}

	currentPage, _ := strconv.Atoi(r.FormValue("page"))
	prevQuery, nextQuery := pageQuerys(r.URL, currentPage)

	tx.Commit()

	app.renderTemplate(w, r, map[string]interface{}{
		"Parts":       partViews,
		"Categories":  categories,
		"Places": places,
		"CurrentPage": currentPage,
		"NextPage":    template.URL(nextQuery),
		"PrevPage":    template.URL(prevQuery),
		"URL":         r.URL,
		"Filter":      filter,
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
	id := path.Base(r.URL.Path)
	http.Redirect(w, r, fmt.Sprintf("/parts/edit/%s", id), http.StatusSeeOther)
}

func (app *Application) EditPartHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		app.UpdatePartHandler(w, r)
		return
	}

	tx := app.DB.MustBegin()
	defer tx.Rollback()

	var err error

	id := path.Base(r.URL.Path)

	partView := new(PartView)
	err = tx.Get(partView, `SELECT * FROM 'part_view' WHERE "id" = ?`, id)
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
	err = tx.Select(&places, `SELECT * FROM 'place'`)
	if err != nil {
		app.Error(w, err)
		return
	}

	categories := []Category{}
	err = tx.Select(&categories, `SELECT * FROM 'category'`)
	if err != nil {
		app.Error(w, err)
		return
	}

	partAmounts := []PartAmount{}
	err = tx.Select(&partAmounts, `SELECT * FROM 'part_amount'
	WHERE "part_id" = ? ORDER BY "timestamp" DESC LIMIT 30`, partView.Id)
	if err != nil {
		app.Error(w, err)
		return
	}

	attachments := []Attachment{}
	err = tx.Select(&attachments, `SELECT * FROM 'attachment'
	WHERE "part_id" = ? ORDER BY "created_at" ASC`, partView.Id)
	if err != nil {
		app.Error(w, err)
		return
	}

	distributorPartViews := []DistributorPartView{}
	err = tx.Select(&distributorPartViews, `SELECT * FROM 'distributor_part_view'
	WHERE "part_id" = ? ORDER BY "name" ASC`, partView.Id)
	if err != nil {
		app.Error(w, err)
		return
	}

	distributors := []Distributor{}
	err = tx.Select(&distributors, `SELECT * FROM 'distributor' ORDER BY 'name' ASC`)
	if err != nil {
		app.Error(w, err)
		return
	}

	app.renderTemplate(w, r, map[string]interface{}{
		"Part":             partView,
		"Categories":       categories,
		"Places":           places,
		"DistributorParts": distributorPartViews,
		"Distributors":     distributors,
		"AmountGraph":      buildPartAmountGraph(partAmounts),
		"Attachments":      attachments,
	}, "EditPart", "Layout")
}

func (app *Application) CreatePartAmountHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.Error(w, err)
		return
	}

	tx := app.DB.MustBegin()
	defer tx.Rollback()

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
	err = tx.Get(part, `SELECT * FROM 'part' WHERE "id" = ?`, partId)
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

	err = partAmount.Save(tx)
	if err != nil {
		app.Error(w, err)
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
		app.Error(w, err)
		return
	}

	tx := app.DB.MustBegin()
	defer tx.Rollback()

	part := new(Part)
	err = tx.Get(part, `SELECT * FROM 'part' WHERE "id" = ?`, id)
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

	err = part.Save(tx)
	if err != nil {
		app.Error(w, err)
		return
	}

	tx.Commit()

	http.Redirect(w, r, fmt.Sprintf("/parts/edit/%d", id), http.StatusFound)
}

func (app *Application) CreatePartHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.Error(w, err)
		return
	}

	tx, err := app.DB.Begin()

	next := r.PostForm.Get("next")
	if next == "" {
		next = "list"
	}

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

	switch next {
	default:
		fallthrough
	case "list":
		http.Redirect(w, r, "/parts", http.StatusSeeOther)
	case "new":
		http.Redirect(w, r, "/parts/new", http.StatusSeeOther)
	case "show":
		http.Redirect(w, r, fmt.Sprintf("/parts/edit/%d", part.Id), http.StatusFound)
	}
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
