-- +goose Up
CREATE TABLE posts(
	id TEXT PRIMARY KEY,
	created_at DATETIME NOT NULL,
	updated_at DATETIME NOT NULL,
	title TEXT NOT NULL,
	url TEXT UNIQUE NOT NULL,
	description TEXT,
	published_at DATETIME,
	feed_id TEXT NOT NULL,
	CONSTRAINT fk_feeds
	FOREIGN KEY (feed_id)
	REFERENCES feeds(id)
	ON DELETE CASCADE
);

-- +goose Down
DROP TABLE posts;
