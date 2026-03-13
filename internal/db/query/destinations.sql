-- name: CreateDestination :one
INSERT INTO destinations (
  name, country, city, description, image_url
) VALUES (
  $1, $2, $3, $4, $5
) RETURNING *;

-- name: GetDestination :one
SELECT * FROM destinations
WHERE id = $1 LIMIT 1;

-- name: ListDestinations :many
SELECT * FROM destinations
WHERE
  ($1::text = '' OR name ILIKE '%' || $1 || '%' OR country ILIKE '%' || $1 || '%' OR city ILIKE '%' || $1 || '%')
  AND ($2::text = '' OR country = $2)
ORDER BY name
LIMIT $3 OFFSET $4;

-- name: CountDestinations :one
SELECT COUNT(*) FROM destinations
WHERE
  ($1::text = '' OR name ILIKE '%' || $1 || '%' OR country ILIKE '%' || $1 || '%' OR city ILIKE '%' || $1 || '%')
  AND ($2::text = '' OR country = $2);

-- name: UpdateDestination :one
UPDATE destinations
SET
  name = COALESCE(sqlc.narg(name), name),
  country = COALESCE(sqlc.narg(country), country),
  city = COALESCE(sqlc.narg(city), city),
  description = COALESCE(sqlc.narg(description), description),
  image_url = COALESCE(sqlc.narg(image_url), image_url),
  updated_at = now()
WHERE id = $1
RETURNING *;
