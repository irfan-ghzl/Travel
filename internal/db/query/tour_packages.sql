-- name: CreateTourPackage :one
INSERT INTO tour_packages (
  title, description, destination_id, price, duration_days,
  max_participants, min_participants, category, image_url
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9
) RETURNING *;

-- name: GetTourPackage :one
SELECT * FROM tour_packages
WHERE id = $1 LIMIT 1;

-- name: ListTourPackages :many
SELECT tp.* FROM tour_packages tp
WHERE
  tp.is_active = true
  AND ($1::text = '' OR tp.title ILIKE '%' || $1 || '%' OR tp.description ILIKE '%' || $1 || '%')
  AND ($2::text = '' OR tp.category = $2)
  AND ($3::bigint = 0 OR tp.destination_id = $3)
  AND ($4::decimal = 0 OR tp.price >= $4)
  AND ($5::decimal = 0 OR tp.price <= $5)
ORDER BY tp.created_at DESC
LIMIT $6 OFFSET $7;

-- name: CountTourPackages :one
SELECT COUNT(*) FROM tour_packages tp
WHERE
  tp.is_active = true
  AND ($1::text = '' OR tp.title ILIKE '%' || $1 || '%' OR tp.description ILIKE '%' || $1 || '%')
  AND ($2::text = '' OR tp.category = $2)
  AND ($3::bigint = 0 OR tp.destination_id = $3)
  AND ($4::decimal = 0 OR tp.price >= $4)
  AND ($5::decimal = 0 OR tp.price <= $5);

-- name: UpdateTourPackage :one
UPDATE tour_packages
SET
  title = COALESCE(sqlc.narg(title), title),
  description = COALESCE(sqlc.narg(description), description),
  price = COALESCE(sqlc.narg(price), price),
  duration_days = COALESCE(sqlc.narg(duration_days), duration_days),
  max_participants = COALESCE(sqlc.narg(max_participants), max_participants),
  min_participants = COALESCE(sqlc.narg(min_participants), min_participants),
  category = COALESCE(sqlc.narg(category), category),
  image_url = COALESCE(sqlc.narg(image_url), image_url),
  is_active = COALESCE(sqlc.narg(is_active), is_active),
  updated_at = now()
WHERE id = $1
RETURNING *;

-- name: DeleteTourPackage :exec
UPDATE tour_packages SET is_active = false, updated_at = now() WHERE id = $1;

-- name: CreateTourItinerary :one
INSERT INTO tour_itineraries (tour_package_id, day_number, title, description)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: ListTourItineraries :many
SELECT * FROM tour_itineraries
WHERE tour_package_id = $1
ORDER BY day_number;

-- name: CreateTourFacility :one
INSERT INTO tour_facilities (tour_package_id, name)
VALUES ($1, $2)
RETURNING *;

-- name: ListTourFacilities :many
SELECT * FROM tour_facilities
WHERE tour_package_id = $1
ORDER BY name;

-- name: CreateTourImage :one
INSERT INTO tour_images (tour_package_id, image_url)
VALUES ($1, $2)
RETURNING *;

-- name: ListTourImages :many
SELECT * FROM tour_images
WHERE tour_package_id = $1;
