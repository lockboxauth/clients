-- +migrate Up
CREATE TABLE clients (
	id VARCHAR(36) PRIMARY KEY,
	secret_hash TEXT NOT NULL DEFAULT '',
	secret_scheme VARCHAR(32) NOT NULL DEFAULT '',
	confidential BOOLEAN NOT NULL DEFAULT false,
	created_at TIMESTAMPTZ NOT NULL,
	created_by VARCHAR(64) NOT NULL DEFAULT '',
	created_by_ip VARCHAR(36) NOT NULL DEFAULT ''
);

CREATE TABLE redirect_uris (
	id VARCHAR(36) PRIMARY KEY,
	uri TEXT NOT NULL,
	is_base_uri BOOLEAN NOT NULL DEFAULT false,
	client_id VARCHAR(36) NOT NULL,
	created_at TIMESTAMPTZ NOT NULL,
	created_by VARCHAR(64) NOT NULL DEFAULT '',
	created_by_ip VARCHAR(36) NOT NULL DEFAULT ''
);

-- +migrate Down
DROP TABLE clients;
DROP TABLE redirect_uris;
