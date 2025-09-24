-- +goose Up
PRAGMA foreign_keys = ON;

CREATE TABLE users(
	id TEXT PRIMARY KEY,
	created_at DATETIME NOT NULL,
	updated_at DATETIME NOT NULL,
	name TEXT UNIQUE NOT NULL
);

-- +goose Down
DROP TABLE users;
