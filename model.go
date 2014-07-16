package inventory

import (
	"crypto/rand"
	"crypto/subtle"
	"database/sql"
	"io"
	"mime"
	"net/url"
	"strconv"
	"strings"
	"time"

	"code.google.com/p/go.crypto/scrypt"
	"github.com/fritz0705/inventory/si"
)

type (
	Execer interface {
		Exec(string, ...interface{}) (sql.Result, error)
	}

	Queryer interface {
		Query(string, ...interface{}) (*sql.Rows, error)
	}

	User struct {
		Id           int64        `db:"id"`
		Name         string       `db:"name"`
		Email        string       `db:"email"`
		PasswordHash []byte       `db:"password_hash"`
		PasswordSalt []byte       `db:"password_salt"`
		IsActive     sql.NullBool `db:"is_active"`
		CreatedAt    time.Time    `db:"created_at"`
		UpdatedAt    time.Time    `db:"updated_at"`
	}

	Part struct {
		Id          int64           `db:"id"`
		Name        string          `db:"name"`
		Description sql.NullString  `db:"description"`
		Value       sql.NullFloat64 `db:"value"`
		CategoryId  int64           `db:"category_id"`
		PlaceId     sql.NullInt64   `db:"place_id"`
		OwnerId     sql.NullInt64   `db:"owner_id"`
		ImageId     sql.NullInt64   `db:"image_id"`
		CreatedAt   time.Time       `db:"created_at"`
	}

	PartView struct {
		Part
		CategoryName string         `db:"category_name"`
		Unit         sql.NullString `db:"unit"`
		UnitSymbol   sql.NullString `db:"unit_symbol"`
		PlaceName    sql.NullString `db:"place_name"`
		Amount       int64          `db:"amount"`
		ImageKey     []byte         `db:"image_key"`
	}

	Category struct {
		Id         int64          `db:"id"`
		Name       string         `db:"name"`
		Unit       sql.NullString `db:"unit"`
		UnitSymbol sql.NullString `db:"unit_symbol"`
	}

	PartAmount struct {
		Id        int64     `db:"id"`
		PartId    int64     `db:"part_id"`
		Amount    int64     `db:"amount"`
		Timestamp time.Time `db:"timestamp"`
	}

	Place struct {
		Id   int64  `db:"id"`
		Name string `db:"name"`
	}

	Attachment struct {
		Id        int64     `db:"id"`
		Key       []byte    `db:"key"`
		Name      string    `db:"name"`
		Type      string    `db:"type"`
		CreatedAt time.Time `db:"created_at"`
		PartId    int64     `db:"part_id"`
	}
)

func (a *Attachment) MediaType() string {
	mediaType, _, err := mime.ParseMediaType(a.Type)
	if err != nil {
		return "application"
	}
	return strings.Split(mediaType, "/")[0]
}

func (u *User) SetPassword(password string) {
	var err error

	u.PasswordSalt = make([]byte, 16)
	_, err = rand.Read(u.PasswordSalt)
	if err != nil {
		panic(err)
	}

	u.PasswordHash, err = scrypt.Key([]byte(password), u.PasswordSalt, 16384, 8, 1, 32)
	if err != nil {
		panic(err)
	}
}

func (u *User) CheckPassword(password string) bool {
	passwordHash, err := scrypt.Key([]byte(password), u.PasswordSalt, 16384, 8, 1, 32)
	if err != nil {
		panic(err)
	}
	return subtle.ConstantTimeCompare(passwordHash, u.PasswordHash) == 1
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
	"id" = ?`, u.Name, u.Email, u.PasswordHash, u.PasswordSalt, u.IsActive,
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

func (a *Attachment) Save(db Execer) error {
	if a.Id == 0 {
		// CREATE
		res, err := db.Exec(`INSERT INTO 'attachment' ('key', 'name', 'type', 'created_at',
		'part_id') VALUES (?, ?, ?, ?, ?)`, a.Key, a.Name, a.Type,
			a.CreatedAt, a.PartId)
		if err != nil {
			return err
		}
		a.Id, err = res.LastInsertId()
		return err
	}

	// UPDATE
	_, err := db.Exec(`UPDATE 'attachment' SET 'key' = ?, 'name' = ?, 'type' = ?,
	'created_at' = ?, 'part_id' = ? WHERE "id" = ?`, a.Key, a.Name, a.Type,
		a.CreatedAt, a.PartId)

	return err
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
		'category_id', 'owner_id', 'place_id', 'created_at', 'image_id')
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			p.Name, p.Description, p.Value, p.CategoryId, p.OwnerId, p.PlaceId,
			p.CreatedAt, p.ImageId)
		if err != nil {
			return err
		}
		p.Id, err = res.LastInsertId()
		return err
	}

	// UPDATE
	_, err := db.Exec(`UPDATE 'part' SET 'name' = ?, 'description' = ?,
	'value' = ?, 'category_id' = ?, 'owner_id' = ?, 'place_id' = ?,
	'created_at' = ?, 'image_id' = ?
	WHERE "id" = ?`, p.Name, p.Description, p.Value, p.CategoryId, p.OwnerId,
		p.PlaceId, p.CreatedAt, p.ImageId, p.Id)

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
		case "created_at":
			dest[n] = &p.CreatedAt
		}
	}

	return rows.Scan(dest...)
}

func (p *Part) LoadForm(form url.Values) error {
	for key, value := range form {
		switch key {
		case "name":
			p.Name = value[0]
		case "description":
			p.Description = sql.NullString{
				String: value[0],
				Valid:  value[0] != "",
			}
		case "value":
			num, err := si.Parse(value[0])
			if err != nil && err != io.EOF {
				return err
			}
			p.Value = sql.NullFloat64{
				Float64: num.Value(),
				Valid:   value[0] != "",
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
			p.PlaceId = sql.NullInt64{
				Int64: int64(val),
				Valid: val != 0,
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
		case "image_id":
			val, err := strconv.Atoi(value[0])
			if err != nil {
				return err
			}
			p.ImageId = sql.NullInt64{
				Int64: int64(val),
				Valid: val != 0,
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

func (p *Part) Place(db Queryer) (*Place, error) {
	rows, err := db.Query(`SELECT * FROM 'place' WHERE "id" = ?`, p.PlaceId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if !rows.Next() {
		return nil, nil
	}

	place := new(Place)
	err = place.Load(rows)

	return place, err
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

	_, err := db.Exec(`UPDATE 'place' SET 'name' = ? WHERE "id" = ?`, p.Name,
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
