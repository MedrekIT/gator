-- +goose Up
CREATE TABLE feed_follows(
	id TEXT PRIMARY KEY,
	created_at DATETIME NOT NULL,
	updated_at DATETIME NOT NULL,
	user_id TEXT NOT NULL,
	feed_id TEXT NOT NULL,
	CONSTRAINT fk_users
	FOREIGN KEY (user_id)
	REFERENCES users(id)
	ON DELETE CASCADE,
	CONSTRAINT fk_feeds
	FOREIGN KEY (feed_id)
	REFERENCES feeds(id)
	ON DELETE CASCADE,
	UNIQUE (user_id, feed_id)
);

-- +goose Down
DROP TABLE feed_follows;

