package inventory

import (
	"database/sql"
	"time"
)

const (
	PolicyFree = iota
	PolicyRestricted
	PolicyAsk
)

const (
	UserTable     = "user"
	ItemTable     = "item"
	LocationTable = "location"
	CategoryTable = "category"
)

func CreateUserTable(db *sql.DB) error {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS "` + UserTable + `" (
		"id" INTEGER PRIMARY KEY,
		"name" TEXT NOT NULL,
		"email" TEXT NOT NULL,
		"epassword_hash" TEXT NOT NULL,
		"password_salt" TEXT NOT NULL,
		"is_active" BOOLEAN DEFAULT TRUE,
		"is_admin" BOOLEAN DEFAULT FALSE,
		"created_at" TIMESTAMP DEFAULT now(),
		"updated_at" TIMESTAMP DEFAULT now()
	)`)
	return err
}

func CreateItemTable(db *sql.DB) error {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS "` + ItemTable + `" (
		"id" INTEGER PRIMARY KEY,
		"name" TEXT NOT NULL,
		"description" TEXT,
		"location_id" INTEGER,
		"category_id" INTEGER,
		"owner_id" INTEGER,
		"policy" INTEGER,
		"amount" INTEGER,
		"created_at" TIMESTAMP DEFAULT now(),
		"updated_at" TIMESTAMP DEFAULT now()
	)`)
	return err
}

func CreateLocationTable(db *sql.DB) error {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS "` + LocationTable + `" (
		"id" INTEGER PRIMARY KEY,
		"name" TEXT NOT NULL
	)`)
	return err
}

func CreateCategoryTable(db *sql.DB) error {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS "` + CategoryTable + `" (
		"id" INTEGER PRIMARY KEY,
		"name" TEXT NOT NULL,
		"parent_id" INTEGER
	)`)
	return err
}

type User struct {
	Id           int
	Name         string
	Email        string
	PasswordHash []byte
	PasswordSalt []byte
	IsActive     bool
	IsAdmin      bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type Item struct {
	Id          int
	Name        string
	Description string
	LocationId  int
	CategoryId  int
	OwnerId     int
	Policy      int
	Amount      int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Location struct {
	Id   int
	Name string
}

type Category struct {
	Id       int
	Name     string
	ParentId int
}

func (i Item) Location(app Application) (*Location, error) {
	rows, err := app.Database.Query(`SELECT * FROM "locations"
	WHERE 'id' = %s LIMIT 1`, i.LocationId)
	if err != nil {
		return nil, err
	}
	if rows.Next() {
	}
	return nil, nil
}

func (i Item) Category(app Application) *Category {
	return nil
}

func (i Item) Owner(app Application) *User {
	return nil
}

func (c Category) Parent(app Application) *Category {
	return nil
}

func (a Application) InitDb() error {
	// TODO(fritz): Implement migrate package to support a migration process
	for _, f := range []func(*sql.DB) error{
		CreateUserTable,
		CreateLocationTable,
		CreateCategoryTable,
		CreateItemTable,
	} {
		err := f(a.Database)
		if err != nil {
			return err
		}
	}
	return nil
}
