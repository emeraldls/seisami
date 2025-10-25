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
INSERT INTO columns (id, board_id, name, position)
VALUES (?, ?, ?, ?)
RETURNING *;

-- name: UpdateColumn :one
UPDATE columns
SET "name" = ?,
    updated_at = CURRENT_TIMESTAMP
WHERE id = ?
RETURNING *;

-- name: DeleteColumn :exec
DELETE FROM columns
WHERE id = ?;

-- 
-- Transcriptions Functionality
--

-- name: GetTranscription :one
SELECT * FROM transcriptions
WHERE id = ?
LIMIT 1;

-- name: ListTranscriptionsByBoard :many
SELECT * FROM transcriptions
WHERE board_id = ?
ORDER BY created_at DESC;

-- name: ListAllTranscriptions :many
SELECT * FROM transcriptions
ORDER BY created_at DESC
LIMIT ? OFFSET ?;

-- name: CreateTranscription :one
INSERT INTO transcriptions (id, board_id, transcription, recording_path, intent, assistant_response)
VALUES (?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: UpdateTranscriptionIntent :one
UPDATE transcriptions
SET intent = ?,
    updated_at = CURRENT_TIMESTAMP
WHERE id = ?
RETURNING *;

-- name: UpdateTranscriptionResponse :one
UPDATE transcriptions
SET assistant_response = ?,
    updated_at = CURRENT_TIMESTAMP
WHERE id = ?
RETURNING *;

-- name: DeleteTranscription :exec
DELETE FROM transcriptions
WHERE id = ?;


-- name: GetTranscriptionByRecordingPath :one
SELECT * FROM transcriptions
where recording_path = ? AND board_id = ?;


-- 
-- Settings Functionality
--

-- name: GetSettings :one
SELECT * FROM settings
WHERE id = 1
LIMIT 1;

-- name: CreateSettings :one
INSERT INTO settings (id, transcription_method, whisper_binary_path, whisper_model_path, openai_api_key)
VALUES (1, ?, ?, ?, ?)
RETURNING *;

-- name: UpdateSettings :one
UPDATE settings
SET transcription_method = ?,
    whisper_binary_path = ?,
    whisper_model_path = ?,
    openai_api_key = ?,
    updated_at = CURRENT_TIMESTAMP
WHERE id = 1
RETURNING *;


-- 
-- Cards Functionality
--

-- name: GetCard :one
SELECT * FROM cards
WHERE id = ?
LIMIT 1;

-- name: ListCardsByColumn :many
SELECT * FROM cards
WHERE column_id = ?
ORDER BY created_at ASC;

-- name: CreateCard :one
INSERT INTO cards (id, column_id, title, description, attachments)
VALUES (?, ?, ?, ?, ?)
RETURNING *;

-- name: UpdateCard :one
UPDATE cards
SET title = ?,
    description = ?,
    attachments = ?,
    updated_at = CURRENT_TIMESTAMP
WHERE id = ?
RETURNING *;

-- name: UpdateCardColumn :one
UPDATE cards
SET column_id = ?
WHERE id = ?
RETURNING *;

-- name: DeleteCard :exec
DELETE FROM cards
WHERE id = ?;


-- name: SearchColumnsByBoardAndName :many
SELECT *
FROM "columns"
WHERE board_id = ?
  AND name LIKE '%' || ? || '%' COLLATE NOCASE;

-- 
-- Export/Sync Functionality
--

-- name: ListAllColumns :many
SELECT * FROM columns
ORDER BY created_at ASC;

-- name: ListAllCards :many
SELECT * FROM cards
ORDER BY created_at ASC;

-- name: CreateOperation :one
INSERT INTO operations (id, table_name, record_id, operation_type, payload)
VALUES (?, ?, ?, ?, ?)
RETURNING *;

--  GetAllOperations The one below is faster & better
-- SELECT * FROM operations
-- WHERE created_at > (SELECT last_synced_at FROM sync_state WHERE sync_state."table_name" = ?)
-- ORDER BY created_at ASC;

-- name: GetAllOperations :many
SELECT o.*
FROM operations o
JOIN (
    SELECT record_id, MAX(created_at) AS max_created_at
    FROM operations
    WHERE created_at > (
        SELECT COALESCE(last_synced_at, 0)
        FROM sync_state
        WHERE sync_state."table_name" = ?
    )
    AND sync_state."table_name" = ?
    GROUP BY record_id
) latest
ON o.record_id = latest.record_id
AND o.created_at = latest.max_created_at
AND o.table_name = ?
ORDER BY o.created_at ASC;


-- name: UpsertSyncState :exec
INSERT INTO sync_state (table_name, last_synced_at, last_synced_op_id)
VALUES (?, ?, ?)
ON CONFLICT(table_name)
DO UPDATE SET
  last_synced_at = excluded.last_synced_at,
  last_synced_op_id = excluded.last_synced_op_id;

-- name: GetSyncState :one
SELECT *
FROM sync_state
WHERE table_name = ?
LIMIT 1;

-- name: UpdateSyncState :exec
UPDATE sync_state
SET last_synced_at = ?, last_synced_op_id = ?
WHERE table_name = ?;
