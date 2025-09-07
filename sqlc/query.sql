-- 
-- Board Functionality
--

-- name: GetBoard :one
SELECT * FROM boards
WHERE id = ?
LIMIT 1;

-- name: ListBoards :many
SELECT * FROM boards
ORDER BY created_at ASC
LIMIT ? OFFSET ?;

-- name: CreateBoard :one
INSERT INTO boards (id, name)
VALUES (?, ?)
RETURNING *;

-- name: UpdateBoard :one
UPDATE boards
SET name = ?,
    updated_at = CURRENT_TIMESTAMP
WHERE id = ?
RETURNING *;

-- name: DeleteBoard :exec
DELETE FROM boards
WHERE id = ?;

-- 
-- Columns Functionality
--

-- name: GetColumn :one
SELECT * FROM columns
WHERE id = ?
LIMIT 1;

-- name: ListColumnsByBoard :many
SELECT * FROM columns
WHERE board_id = ?
ORDER BY created_at ASC;

-- name: CreateColumn :one
INSERT INTO columns (id, board_id, title, position)
VALUES (?, ?, ?, ?)
RETURNING *;

-- name: UpdateColumn :one
UPDATE columns
SET title = ?,
    updated_at = CURRENT_TIMESTAMP
WHERE id = ?
RETURNING *;

-- name: DeleteColumn :exec
DELETE FROM columns
WHERE id = ?;


-- 
-- Tickets Functionality
--

-- name: GetTicket :one
SELECT * FROM tickets
WHERE id = ?
LIMIT 1;

-- name: ListTicketsByColumn :many
SELECT * FROM tickets
WHERE column_id = ?
ORDER BY created_at ASC;

-- name: CreateTicket :one
INSERT INTO tickets (id, column_id, title, description, assignee_id, story_points, pr_link, ticket_type)
VALUES (?, ?, ?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: UpdateTicket :one
UPDATE tickets
SET title = ?,
    description = ?,
    assignee_id = ?,
    story_points = ?,
    pr_link = ?,
    ticket_type = ?,
    updated_at = CURRENT_TIMESTAMP
WHERE id = ?
RETURNING *;

-- name: UpdateTicketColumn :one
UPDATE tickets
SET column_id = ?
WHERE id = ?
RETURNING *;

-- name: DeleteTicket :exec
DELETE FROM tickets
WHERE id = ?;
