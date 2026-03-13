-- name: CreateTourSchedule :one
INSERT INTO tour_schedules (
  tour_package_id, start_date, end_date, available_slots
) VALUES (
  $1, $2, $3, $4
) RETURNING *;

-- name: GetTourSchedule :one
SELECT * FROM tour_schedules
WHERE id = $1 LIMIT 1;

-- name: ListTourSchedules :many
SELECT * FROM tour_schedules
WHERE
  tour_package_id = $1
  AND ($2::date = '0001-01-01' OR start_date >= $2)
  AND ($3::date = '0001-01-01' OR end_date <= $3)
ORDER BY start_date;

-- name: UpdateTourScheduleSlots :one
UPDATE tour_schedules
SET available_slots = available_slots - $2, updated_at = now()
WHERE id = $1 AND available_slots >= $2
RETURNING *;

-- name: UpdateTourScheduleStatus :one
UPDATE tour_schedules
SET status = $2, updated_at = now()
WHERE id = $1
RETURNING *;
