-- +goose Up
ALTER TABLE users
ADD COLUMN hashed_password TEXT NOT NULL DEFAULT 'unset';

UPDATE users
SET hashed_password = 'unset';

-- +goose Down
ALTER TABLE users
DROP COLUMN hashed_password;