package inventory

import (
	"database/sql"
	"net/url"
	"strconv"
	"time"

	"github.com/fritz0705/inventory/si"
)

type Execer interface {
	Exec(string, ...interface{}) (sql.Result, error)
}

type Queryer interface {
	Query(string, ...interface{}) (*sql.Rows, error)
}

type User struct {
	Id           int64
	Name         string
	Email        string
	PasswordHash []byte
	PasswordSalt []byte
	IsActive     sql.NullBool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type Part struct {
	Id          int64
	Name        string
	Description sql.NullString
	Value       sql.NullFloat64
	CategoryId  int64
	PlaceId  sql.NullInt64
	OwnerId     sql.NullInt64
}

type Category struct {
	Id         int64
	Name       string
	Unit       sql.NullString
	UnitSymbol sql.NullString
}

type PartAmount struct {
	Id        int64
	PartId    int64
	Amount    int64
	Timestamp time.Time
}

type Place struct {
	Id   int64
	Name string
}

func LoadCategories(db Queryer, query string, p ...interface{}) ([]*Category, error) {
	rows, err := db.Query(query, p...)
	if err != nil {
		return nil, err
	}

	categories := make([]*Category, 0)
	for rows.Next() {
		category := new(Category)
		err := category.Load(rows)
		if err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}

	return categories, nil
}

func LoadPlaces(db Queryer, query string, p ...interface{}) ([]*Place, error) {
	rows, err := db.Query(query, p...)
	if err != nil {
		return nil, err
	}

	places := make([]*Place, 0)
	for rows.Next() {
		place := new(Place)
		err := place.Load(rows)
		if err != nil {
			return nil, err
		}
		places = append(places, place)
	}

	return places, nil
}

func LoadPartAmounts(db Queryer, query string, p ...interface{}) ([]*PartAmount, error) {
	rows, err := db.Query(query, p...)
	if err != nil {
		return nil, err
	}

	partAmounts := make([]*PartAmount, 0)
	for rows.Next() {
		partAmount := new(PartAmount)
		err := partAmount.Load(rows)
		if err != nil {
			return nil, err
		}
		partAmounts = append(partAmounts, partAmount)
	}

	return partAmounts, nil
}

func FindPart(db Queryer, id int64) (*Part, error) {
	rows, err := db.Query(`SELECT * FROM 'part' WHERE "id" = ? LIMIT 1`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if !rows.Next() {
		return nil, nil
	}

	part := new(Part)
	err = part.Load(rows)

	return part, err
}

func (u *User) SetPassword(password string) {
	u.PasswordHash = []byte(password)
}

func (u *User) CheckPassword(password string) bool {
	return password == string(u.PasswordHash)
}

func (u *User) Save(db Execer) error {
	if u.Id == 0 {
		// CREATE
		res, err := db.Exec(`INSERT INTO 'user' ('name', 'email', 'password_hash',
		'password_salt', 'is_active', 'created_at', 'updated_at') VALUES (?, ?, ?, ?, ?, ?, ?);`,
			u.Name, u.Email, u.PasswordHash, u.PasswordSalt, u.IsActive, u.CreatedAt, u.UpdatedAt)
		if err != nil {
			return err
		}
		u.Id, err = res.LastInsertId()
		return err
	}

	// UPDATE
	_, err := db.Exec(`UPDATE 'user' SET 'name' = ?, 'email' = ?, 'password_hash' = ?,
	'password_salt' = ?, 'is_active' = ?, 'created_at' = ?, 'updated_at' = ? WHERE
	'id' = ?`, u.Name, u.Email, u.PasswordHash, u.PasswordSalt, u.IsActive,
		u.CreatedAt, u.UpdatedAt, u.Id)

	return err
}

func (u *User) Load(rows *sql.Rows) error {
	columns, err := rows.Columns()
	if err != nil {
		return err
	}
	dest := make([]interface{}, len(columns))
	for n, col := range columns {
		switch col {
		case "id":
			dest[n] = &u.Id
		case "name":
			dest[n] = &u.Name
		case "email":
			dest[n] = &u.Email
		case "password_hash":
			dest[n] = &u.PasswordHash
		case "password_salt":
			dest[n] = &u.PasswordSalt
		case "is_active":
			dest[n] = &u.IsActive
		case "created_at":
			dest[n] = &u.CreatedAt
		case "updated_at":
			dest[n] = &u.UpdatedAt
		}
	}

	return rows.Scan(dest...)
}

func LoadParts(rows *sql.Rows) ([]*Part, error) {
	parts := make([]*Part, 0)
	for rows.Next() {
		part := new(Part)
		err := part.Load(rows)
		if err != nil {
			return nil, err
		}
		parts = append(parts, part)
	}

	return parts, nil
}

func (p *Part) Save(db Execer) error {
	if p.Id == 0 {
		// CREATE
		res, err := db.Exec(`INSERT INTO 'part' ('name', 'description', 'value',
		'category_id', 'owner_id', 'place_id') VALUES (?, ?, ?, ?, ?, ?)`,
			p.Name, p.Description, p.Value, p.CategoryId, p.OwnerId, p.PlaceId)
		if err != nil {
			return err
		}
		p.Id, err = res.LastInsertId()
		return err
	}

	// UPDATE
	_, err := db.Exec(`UPDATE 'part' SET 'name' = ?, 'description' = ?,
	'value' = ?, 'category_id' = ?, 'owner_id' = ?, 'place_id' = ? WHERE 'id' = ?`, p.Name,
		p.Description, p.Value, p.CategoryId, p.OwnerId, p.PlaceId, p.Id)

	return err
}

func (p *Part) Load(rows *sql.Rows) error {
	columns, err := rows.Columns()
	if err != nil {
		return err
	}
	dest := make([]interface{}, len(columns))
	for n, col := range columns {
		switch col {
		case "id":
			dest[n] = &p.Id
		case "name":
			dest[n] = &p.Name
		case "description":
			dest[n] = &p.Description
		case "value":
			dest[n] = &p.Value
		case "category_id":
			dest[n] = &p.CategoryId
		case "owner_id":
			dest[n] = &p.OwnerId
		case "place_id":
			dest[n] = &p.PlaceId
		}
	}

	return rows.Scan(dest...)
}

func (p *Part) LoadForm(form url.Values) error {
	for key, value := range form {
		if value[0] == "" {
			continue
		}
		switch key {
		case "name":
			p.Name = value[0]
		case "description":
			p.Description = sql.NullString{
				String: value[0],
				Valid:  true,
			}
		case "value":
			num, err := si.Parse(value[0])
			if err != nil {
				return err
			}
			p.Value = sql.NullFloat64{
				Float64: num.Value(),
				Valid:   true,
			}
		case "category":
			val, err := strconv.Atoi(value[0])
			if err != nil {
				return err
			}
			p.CategoryId = int64(val)
		case "place":
			val, err := strconv.Atoi(value[0])
			if err != nil {
				return err
			}
			if val != 0 {
				p.PlaceId = sql.NullInt64{
					Int64: int64(val),
					Valid: true,
				}
			}
		case "owner":
			val, err := strconv.Atoi(value[0])
			if err != nil {
				return err
			}
			p.OwnerId = sql.NullInt64{
				Int64: int64(val),
				Valid: true,
			}
		}
	}
	return nil
}

func (p *Part) LatestAmount(db Queryer) (*PartAmount, error) {
	rows, err := db.Query(`SELECT * FROM 'part_amount' WHERE "part_id" = ? ORDER BY "timestamp" DESC LIMIT 1`, p.Id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if !rows.Next() {
		return nil, nil
	}

	amount := new(PartAmount)
	err = amount.Load(rows)

	return amount, err
}

func (p *Part) Category(db Queryer) (*Category, error) {
	rows, err := db.Query(`SELECT * FROM 'category' WHERE "id" = ?`, p.CategoryId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if !rows.Next() {
		return nil, nil
	}

	category := new(Category)
	err = category.Load(rows)

	return category, err
}

func (c *Category) Save(db Execer) error {
	if c.Id == 0 {
		res, err := db.Exec(`INSERT INTO 'category' ('name', 'unit', 'unit_symbol') VALUES (?, ?, ?)`,
			c.Name, c.Unit, c.UnitSymbol)
		if err != nil {
			return err
		}
		c.Id, err = res.LastInsertId()
		return err
	}

	// UPDATE
	_, err := db.Exec(`UPDATE 'category' SET 'name' = ?, 'unit' = ?, 'unit_symbol' = ? WHERE "id" = ?`,
		c.Name, c.Unit, c.UnitSymbol, c.Id)

	return err
}

func (c *Category) Load(rows *sql.Rows) error {
	columns, err := rows.Columns()
	if err != nil {
		return err
	}
	dest := make([]interface{}, len(columns))
	for n, col := range columns {
		switch col {
		case "id":
			dest[n] = &c.Id
		case "name":
			dest[n] = &c.Name
		case "unit":
			dest[n] = &c.Unit
		case "unit_symbol":
			dest[n] = &c.UnitSymbol
		}
	}

	return rows.Scan(dest...)
}

func (c *Category) LoadForm(form url.Values) error {
	for key, value := range form {
		if value[0] == "" {
			continue
		}
		switch key {
		case "name":
			c.Name = value[0]
		case "unit":
			c.Unit = sql.NullString{
				String: value[0],
				Valid:  true,
			}
		case "unit_symbol":
			c.UnitSymbol = sql.NullString{
				String: value[0],
				Valid:  true,
			}
		}
	}

	return nil
}

func (p *PartAmount) Save(db Execer) error {
	if p.Id == 0 {
		res, err := db.Exec(`INSERT INTO 'part_amount' ('part_id', 'amount', 'timestamp')
		VALUES (?, ?, ?)`, p.PartId, p.Amount, p.Timestamp)
		if err != nil {
			return err
		}
		p.Id, err = res.LastInsertId()
		return err
	}

	// UPDATE
	_, err := db.Exec(`UPDATE 'part_amount' SET 'part_id' = ?, 'amount' = ?,
	'timestamp' = ? WHERE 'id' = ?`, p.PartId, p.Amount, p.Timestamp, p.Id)

	return err
}

func (p *PartAmount) Load(rows *sql.Rows) error {
	columns, err := rows.Columns()
	if err != nil {
		return err
	}
	dest := make([]interface{}, len(columns))
	for n, col := range columns {
		switch col {
		case "id":
			dest[n] = &p.Id
		case "part_id":
			dest[n] = &p.PartId
		case "amount":
			dest[n] = &p.Amount
		case "timestamp":
			dest[n] = &p.Timestamp
		}
	}

	return rows.Scan(dest...)
}

func (p *Place) Save(db Execer) error {
	if p.Id == 0 {
		res, err := db.Exec(`INSERT INTO 'place' ('name') VALUES (?)`, p.Name)
		if err != nil {
			return err
		}
		p.Id, err = res.LastInsertId()
		return err
	}

	_, err := db.Exec(`UPDATE 'place' SET 'name' = ? WHERE 'id' = ?`, p.Name,
		p.Id)

	return err
}

func (l *Place) Load(rows *sql.Rows) error {
	columns, err := rows.Columns()
	if err != nil {
		return err
	}
	dest := make([]interface{}, len(columns))
	for n, col := range columns {
		switch col {
		case "id":
			dest[n] = &l.Id
		case "name":
			dest[n] = &l.Name
		}
	}

	return rows.Scan(dest...)
}

func (l *Place) LoadForm(form url.Values) error {
	for key, value := range form {
		if value[0] == "" {
			continue
		}

		switch key {
		case "name":
			l.Name = value[0]
		}
	}

	return nil
}
