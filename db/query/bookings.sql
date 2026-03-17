-- name: CreateBooking :one
INSERT INTO bookings (
  booking_code, user_id, tour_package_id, tour_schedule_id,
  travel_date, num_participants, total_price, notes
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8
) RETURNING *;

-- name: GetBooking :one
SELECT * FROM bookings
WHERE id = $1 LIMIT 1;

-- name: GetBookingByCode :one
SELECT * FROM bookings
WHERE booking_code = $1 LIMIT 1;

-- name: ListBookingsByUser :many
SELECT * FROM bookings
WHERE
  user_id = $1
  AND ($2::text = '' OR status = $2)
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;

-- name: CountBookingsByUser :one
SELECT COUNT(*) FROM bookings
WHERE
  user_id = $1
  AND ($2::text = '' OR status = $2);

-- name: ListAllBookings :many
SELECT * FROM bookings
WHERE ($1::text = '' OR status = $1)
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountAllBookings :one
SELECT COUNT(*) FROM bookings
WHERE ($1::text = '' OR status = $1);

-- name: UpdateBookingStatus :one
UPDATE bookings
SET status = $2, updated_at = now()
WHERE id = $1
RETURNING *;

-- name: CreateBookingParticipant :one
INSERT INTO booking_participants (
  booking_id, name, id_card_number, date_of_birth
) VALUES (
  $1, $2, $3, $4
) RETURNING *;

-- name: ListBookingParticipants :many
SELECT * FROM booking_participants
WHERE booking_id = $1;
