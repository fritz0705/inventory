
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
CREATE TABLE '_category' (
	'id' INTEGER PRIMARY KEY,
	'name' TEXT NOT NULL,
	'unit' TEXT,
	'unit_symbol' TEXT,
	'parent_id' INTEGER,
	FOREIGN KEY('parent_id') REFERENCES 'category'('id') ON DELETE SET NULL
);

INSERT INTO '_category' SELECT *, NULL FROM 'category';
DROP TABLE 'category';
ALTER TABLE '_category' RENAME TO 'category';
CREATE INDEX 'category_idx_parent_id' ON 'category'('parent_id');
PRAGMA foreign_key_check;


-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back

