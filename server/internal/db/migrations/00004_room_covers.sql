-- +goose Up
ALTER TABLE rooms ADD COLUMN cover_image BYTEA,
 ADD COLUMN cover_thumbnail BYTEA,
 ADD COLUMN cover_alt_text TEXT NOT NULL DEFAULT '',
 ADD COLUMN cover_version BIGINT NOT NULL DEFAULT 0,
 ADD COLUMN cover_is_demo BOOLEAN NOT NULL DEFAULT FALSE;

-- +goose Down
ALTER TABLE rooms DROP COLUMN cover_image, DROP COLUMN cover_thumbnail,
 DROP COLUMN cover_alt_text, DROP COLUMN cover_version, DROP COLUMN cover_is_demo;
