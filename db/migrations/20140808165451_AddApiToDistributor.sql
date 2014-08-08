
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied

ALTER TABLE 'distributor' ADD COLUMN 'api' TEXT;
DROP VIEW 'distributor_part_view';
CREATE VIEW 'distributor_part_view' AS SELECT 'distributor_part'.*,
	'distributor'.'name' AS 'name',
	'distributor'.'url' AS 'url',
	'distributor'.'homepage' AS 'homepage',
	'distributor'.'api' AS 'api'
	FROM 'distributor_part'
	LEFT JOIN 'distributor' ON 'distributor'.'id' = 'distributor_part'.'distributor_id';


-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back

