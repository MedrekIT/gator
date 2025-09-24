-- name: CreateFeedFollow :one
INSERT INTO feed_follows (id, created_at, updated_at, user_id, feed_id)
VALUES (
	?1,
	?2,
	?3,
	?4,
	?5
)
RETURNING
(SELECT name AS user_name
FROM users
WHERE feed_follows.user_id = users.id),
(SELECT feeds.name AS feed_name
FROM feeds
WHERE feed_follows.feed_id = feeds.id),
*;

-- name: GetFeedFollowsForUser :many
SELECT users.name AS user_name,
feeds.name AS feed_name
FROM feed_follows
INNER JOIN users
ON feed_follows.user_id = users.id
INNER JOIN feeds
ON feed_follows.feed_id = feeds.id
WHERE users.id = ?1;

-- name: DeleteFeedFollow :exec
DELETE FROM feed_follows
WHERE user_id = ?1 AND feed_id = ?2;
