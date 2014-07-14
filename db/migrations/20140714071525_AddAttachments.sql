
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
CREATE TABLE IF NOT EXISTS 'attachment' (
	'id' INTEGER PRIMARY KEY,
	'key' BLOB NOT NULL,
	'name' TEXT NOT NULL,
	'type' TEXT NOT NULL,
	'created_at' DATETIME,
	'part_id' INTEGER NOT NULL,
	FOREIGN KEY('part_id') REFERENCES 'part'('id') ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS 'attachment_idx_part_id' ON 'attachment' (
	'part_id'
);

CREATE INDEX IF NOT EXISTS 'attachment_idx_key' ON 'attachment'('key');

ALTER TABLE 'part' ADD COLUMN 'image_id' INTEGER;

CREATE INDEX IF NOT EXISTS 'part_idx_image_id' ON 'part' ('image_id');

DROP VIEW 'part_view';

CREATE VIEW IF NOT EXISTS 'part_view' AS SELECT 'part'.*,
	'category'."name" AS 'category_name',
	'category'."unit" AS 'unit',
	'category'."unit_symbol" AS 'unit_symbol',
	'place'."name" AS 'place_name',
	"part_amount" AS 'amount',
	'attachment'.'key' AS 'image_key'
	FROM 'part'
	LEFT JOIN 'category' ON 'category'.'id' = 'part'.'category_id'
	LEFT JOIN 'place' ON 'place'.'id' = 'part'.'place_id'
	LEFT JOIN 'attachment' ON 'attachment'.'id' = 'part'.'image_id'
	LEFT JOIN (SELECT "amount" AS 'part_amount',
		"part_id" AS 'part_amount_part_id' FROM 'part_amount'
		GROUP BY "part_amount_part_id"
		ORDER BY "timestamp" DESC) ON "part_amount_part_id" = 'part'.'id';

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back

