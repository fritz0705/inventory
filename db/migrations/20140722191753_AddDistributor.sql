
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied

CREATE TABLE IF NOT EXISTS 'distributor' (
	'id' INTEGER PRIMARY KEY,
	'name' TEXT NOT NULL,
	'url' TEXT NOT NULL,
	'homepage' TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS 'distributor_part' (
	'id' INTEGER PRIMARY KEY,
	'distributor_id' INTEGER NOT NULL,
	'part_id' INTEGER NOT NULL,
	'price' INTEGER,
	'key' TEXT NOT NULL,
	FOREIGN KEY('distributor_id') REFERENCES 'distributor'('id') ON DELETE CASCADE,
	FOREIGN KEY('part_id') REFERENCES 'part'('id') ON DELETE CASCADE
);

CREATE VIEW IF NOT EXISTS 'distributor_part_view' AS SELECT 'distributor_part'.*,
	'distributor'.'name' AS 'name',
	'distributor'.'url' AS 'url',
	'distributor'.'homepage' AS 'homepage'
	FROM 'distributor_part'
	LEFT JOIN 'distributor' ON 'distributor'.'id' = 'distributor_part'.'distributor_id';

CREATE INDEX IF NOT EXISTS 'distributor_part_idx_distributor_id' ON 
	'distributor_part'('distributor_id');
CREATE INDEX IF NOT EXISTS 'distributor_part_idx_part_id' ON
	'distributor_part'('part_id');

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back

DROP TABLE 'distributor';
DROP TABLE 'distributor_part';

