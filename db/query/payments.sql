-- name: CreatePayment :one
INSERT INTO payments (
  booking_id, payment_method, amount, status, payment_token, payment_url, expires_at
) VALUES (
  $1, $2, $3, $4, $5, $6, $7
) RETURNING *;

-- name: GetPaymentByBookingID :one
SELECT * FROM payments
WHERE booking_id = $1 LIMIT 1;

-- name: GetPaymentByID :one
SELECT * FROM payments
WHERE id = $1 LIMIT 1;

-- name: UpdatePaymentStatus :one
UPDATE payments
SET
  status = $2,
  paid_at = CASE WHEN $2 = 'paid' THEN now() ELSE paid_at END,
  updated_at = now()
WHERE id = $1
RETURNING *;

-- name: UpdatePaymentToken :one
UPDATE payments
SET
  payment_token = $2,
  payment_url = $3,
  updated_at = now()
WHERE id = $1
RETURNING *;

-- name: ListPayments :many
SELECT * FROM payments
WHERE ($1::text = '' OR status = $1)
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountPayments :one
SELECT COUNT(*) FROM payments
WHERE ($1::text = '' OR status = $1);
