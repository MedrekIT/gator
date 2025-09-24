-- +goose Up
CREATE TABLE feeds(
	id TEXT PRIMARY KEY,
	created_at DATETIME NOT NULL,
	updated_at DATETIME NOT NULL,
	name TEXT NOT NULL,
	url TEXT UNIQUE NOT NULL,
	user_id TEXT NOT NULL,
	CONSTRAINT fk_users
	FOREIGN KEY (user_id)
	REFERENCES users(id)
	ON DELETE CASCADE
);

-- +goose Down
DROP TABLE feeds;
