package inventory

import (
	"net/http"

	"github.com/fritz0705/inventory/si"
)

type searchQuery struct {
	Unit string
	Value si.Number
	Stock si.Number
	Keywords []string
}

func (s searchQuery) SQL() (query string, args []interface{}) {
	query += `SELECT * FROM 'part_view' WHERE 1=1`

	if s.Unit != "" {
		query += ` AND ("unit" = ? OR "unit_symbol" = ?)`
		args = append(args, s.Unit, s.Unit)
	}

	if s.Value.Value() != 0 {
		query += ` AND "value" = ?`
		args = append(args, s.Value.Value())
	}

	if s.Stock.Value() != 0 {
		query += ` AND "amount" = ?`
		args = append(args, s.Stock.Value())
	}

	for _, kw := range s.Keywords {
		query += ` AND "name" LIKE ?`
		args = append(args, kw)
	}

	return
}

func loadSearchQuery(c chan searchItem) (*searchQuery, error) {
	res := new(searchQuery)
	for item := range c {
		var err error
		switch item.typ {
		case searchItemUnit:
			res.Unit = item.val[1:len(item.val)-1]
		case searchItemNumber:
			res.Value, err = si.Parse(item.val)
		case searchItemText:
			res.Keywords = append(res.Keywords, item.val)
		case searchItemStock:
			res.Stock, err = si.Parse(item.val[1:len(item.val)-1])
		}
		if err != nil {
			return nil, err
		}
	}

	return res, nil
}

func (app *Application) SearchHandler(w http.ResponseWriter, r *http.Request) {
	query := r.FormValue("query")

	_, c := searchLex("search", query)
	sq, err := loadSearchQuery(c)
	if err != nil {
		app.Error(w, err)
		return
	}

	sql, args := sq.SQL()

	res := []PartView{}
	err = app.DB.Select(&res, sql, args...)
	if err != nil {
		app.Error(w, err)
		return
	}

	if len(res) == 1 {
		res := res[0]
		http.Redirect(w, r, fmt.Sprintf("/parts/%d", res.Id), http.StatusFound)
		return
	}

	app.renderTemplate(w, r, map[string]interface{}{
		"Parts": res,
	}, "Search", "Layout")
}

type searchItemType int

const (
	searchItemError searchItemType = iota

	searchItemNumber
	searchItemEOF
	searchItemText
	searchItemUnit
	searchItemStock
)

type searchStateFunc func(*searchLexer) searchStateFunc

type searchItem struct {
	typ searchItemType
	val string
}

func (s searchItem) String() string {
	typ := ""
	switch s.typ {
	case searchItemEOF:
		typ = "EOF"
	case searchItemText:
		typ = "Text"
	case searchItemUnit:
		typ = "Unit"
	case searchItemNumber:
		typ = "Number"
	case searchItemError:
		typ = "Error"
	}
	return typ + "(" + s.val + ")"
}

type searchLexer struct {
	name string
	input string
	start int
	pos int
	width int
	items chan searchItem
}

func searchLex(name, input string) (*searchLexer, chan searchItem) {
	l := &searchLexer{
		name: name,
		input: input,
		items: make(chan searchItem),
	}
	go l.run()
	return l, l.items
}

func (l *searchLexer) run() {
	for state := searchLexAny; state != nil; {
		state = state(l)
	}
	close(l.items)
}

func (l *searchLexer) emit(t searchItemType) {
	l.items <- searchItem{t, l.input[l.start:l.pos]}
	l.start = l.pos
}

func searchLexAny(l *searchLexer) searchStateFunc {
	for l.pos < len(l.input) && l.input[l.pos] <= ' ' {
		l.pos++
	}
	if l.pos >= len(l.input) {
		l.emit(searchItemEOF)
		return nil
	}
	l.start = l.pos

	if l.input[l.pos] == '[' {
		return searchLexUnit
	} else if l.input[l.pos] == '<' {
		return searchLexStock
	}
	return searchLexNumberOrText
}

func searchLexStock(l *searchLexer) searchStateFunc {
	for l.pos < len(l.input) && l.input[l.pos] != '>' {
		l.pos++
	}
	if l.pos >= len(l.input) {
		l.emit(searchItemEOF)
		return nil
	}
	l.pos++
	l.emit(searchItemStock)
	return searchLexAny
}

func searchLexUnit(l *searchLexer) searchStateFunc {
	for l.pos < len(l.input) && l.input[l.pos] != ']' {
		l.pos++
	}
	if l.pos >= len(l.input) {
		l.emit(searchItemEOF)
		return nil
	}
	l.pos++
	l.emit(searchItemUnit)
	return searchLexAny
}

func searchLexText(l *searchLexer) searchStateFunc {
	for {
		if l.pos >= len(l.input) || l.input[l.pos] == ' ' {
			break
		}
		l.pos++
	}
	l.emit(searchItemText)
	if l.pos >= len(l.input) {
		l.emit(searchItemEOF)
		return nil
	}
	return searchLexAny
}

func searchLexNumber(l *searchLexer) searchStateFunc {
	for l.input[l.pos] != ' ' {
		if l.input[l.pos] == '[' {
			l.emit(searchItemNumber)
			return searchLexUnit
		}
		l.pos++
		if l.pos >= len(l.input) {
			break
		}
	}
	l.emit(searchItemNumber)
	return searchLexAny
}

func searchLexNumberOrText(l *searchLexer) searchStateFunc {
	switch l.input[l.pos] {
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		return searchLexNumber
	}
	return searchLexText
}
