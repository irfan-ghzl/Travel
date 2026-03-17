-- name: CreateUser :one
INSERT INTO users (
  email, name, phone, password_hash, google_id, avatar_url, role
) VALUES (
  $1, $2, $3, $4, $5, $6, $7
) RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1 LIMIT 1;

-- name: GetUserByID :one
SELECT * FROM users
WHERE id = $1 LIMIT 1;

-- name: GetUserByGoogleID :one
SELECT * FROM users
WHERE google_id = $1 LIMIT 1;

-- name: UpdateUser :one
UPDATE users
SET
  name = COALESCE(sqlc.narg(name), name),
  phone = COALESCE(sqlc.narg(phone), phone),
  avatar_url = COALESCE(sqlc.narg(avatar_url), avatar_url),
  updated_at = now()
WHERE id = $1
RETURNING *;

-- name: UpdateUserVerified :one
UPDATE users
SET is_verified = true, updated_at = now()
WHERE id = $1
RETURNING *;

-- name: ListUsers :many
SELECT * FROM users
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountUsers :one
SELECT COUNT(*) FROM users;
