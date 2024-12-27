-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens(token, created_at, updated_at, user_id, expires_at)
VALUES ($1, NOW(), NOW(), $2, $3)
RETURNING *;

-- name: GetUserFromRefreshToken :one
SELECT
    users.id AS id,
    users.created_at AS created_at,
    users.updated_at AS updated_at,
    users.email AS email
FROM
    refresh_tokens
        INNER JOIN
    users ON refresh_tokens.user_id = users.id
WHERE
    refresh_tokens.token = $1
  AND refresh_tokens.expires_at > NOW()
  AND refresh_tokens.revoked_at IS NULL;

-- name: RevokeRefreshToken :exec
UPDATE refresh_tokens
SET
    revoked_at = NOW(),
    updated_at = NOW()
WHERE token = $1 AND revoked_at IS NULL;