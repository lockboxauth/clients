-- +migrate Up
ALTER TABLE clients ADD COLUMN name TEXT NOT NULL;

-- +migrate Down
ALTER TABLE clients DROP COLUMN name;
