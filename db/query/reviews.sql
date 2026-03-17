-- name: CreateReview :one
INSERT INTO reviews (
  user_id, tour_package_id, booking_id, rating, comment
) VALUES (
  $1, $2, $3, $4, $5
) RETURNING *;

-- name: GetReview :one
SELECT * FROM reviews
WHERE id = $1 LIMIT 1;

-- name: GetReviewByBooking :one
SELECT * FROM reviews
WHERE user_id = $1 AND booking_id = $2 LIMIT 1;

-- name: ListReviewsByTour :many
SELECT r.*, u.name as user_name, u.avatar_url as user_avatar
FROM reviews r
JOIN users u ON r.user_id = u.id
WHERE r.tour_package_id = $1
ORDER BY r.created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountReviewsByTour :one
SELECT COUNT(*) FROM reviews
WHERE tour_package_id = $1;

-- name: GetAverageRating :one
SELECT COALESCE(AVG(rating), 0)::decimal(3,2) as average_rating
FROM reviews
WHERE tour_package_id = $1;

-- name: DeleteReview :exec
DELETE FROM reviews WHERE id = $1;
