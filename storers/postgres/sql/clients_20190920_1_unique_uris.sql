-- +migrate Up
ALTER TABLE redirect_uris ADD CONSTRAINT redirect_uris_unique_uri UNIQUE(uri);

-- +migrate Down
ALTER TABLE redirect_uris DROP CONSTRAINT redirect_uris_unique_uri;
