CREATE TABLE IF NOT EXISTS 'user' (
	'id' INTEGER PRIMARY KEY NOT NULL,
	'name' TEXT NOT NULL,
	'email' TEXT NOT NULL UNIQUE,
	'password_hash' BLOB NOT NULL,
	'password_salt' BLOB NOT NULL,
	'is_active' BOOLEAN DEFAULT TRUE,
	'created_at' DATETIME,
	'updated_at' DATETIME
);

CREATE INDEX IF NOT EXISTS 'user_idx_email' ON 'user' (
	'email'
);

CREATE TABLE IF NOT EXISTS 'category' (
	'id' INTEGER PRIMARY KEY NOT NULL,
	'name' TEXT NOT NULL,
	'unit' TEXT,
	'unit_symbol' TEXT
);

CREATE TABLE IF NOT EXISTS 'part' (
	'id' INTEGER PRIMARY KEY,
	'name' TEXT NOT NULL,
	'description' TEXT,
	'value' INTEGER,
	'category_id' INTEGER NOT NULL,
	'place_id' INTEGER,
	'owner_id' INTEGER,
	'created_at' DATETIME,
	FOREIGN KEY('category_id') REFERENCES 'category'('id'),
	FOREIGN KEY('place_id') REFERENCES 'place'('id'),
	FOREIGN KEY('owner_id') REFERENCES 'user'('id') ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS 'part_idx_name' ON 'part' (
	'name'
);

CREATE INDEX IF NOT EXISTS 'part_idx_value' ON 'part' (
	'value'
);

CREATE TABLE IF NOT EXISTS 'part_amount' (
	'id' INTEGER PRIMARY KEY,
	'part_id' INTEGER NOT NULL,
	'amount' INTEGER NOT NULL,
	'timestamp' DATETIME,
	FOREIGN KEY('part_id') REFERENCES 'part'('id') ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS 'part_amount_idx_timestamp' ON 'part_amount' (
	'timestamp'
);

CREATE TABLE IF NOT EXISTS 'place' (
	'id' INTEGER PRIMARY KEY,
	'name' TEXT NOT NULL
);

CREATE VIEW IF NOT EXISTS 'part_view' AS SELECT 'part'."id" AS 'id',
	'part'."name" AS 'name',
	'part'."value" AS 'value',
	'part'."description" AS 'description',
	'part'."category_id" AS 'category_id',
	'category'."name" AS 'category_name',
	'category'."unit" AS 'unit',
	'category'."unit_symbol" AS 'unit_symbol',
	'part'."place_id" AS 'place_id',
	'place'."name" AS 'place_name',
	"part_amount" AS 'amount'
	FROM 'part'
	LEFT JOIN 'category' ON 'category'.'id' = 'part'.'category_id'
	LEFT JOIN 'place' ON 'place'.'id' = 'part'.'place_id'
	LEFT JOIN (SELECT "amount" AS 'part_amount',
		"part_id" AS 'part_amount_part_id' FROM 'part_amount'
		GROUP BY "part_amount_part_id"
		ORDER BY "timestamp" DESC) ON "part_amount_part_id" = 'part'.'id';
