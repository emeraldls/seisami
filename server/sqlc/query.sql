-- name: CreateUser :one
INSERT INTO users (id, email, password_hash)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetUserByEmail :one
SELECT *
FROM users
WHERE email = $1;

-- name: UpdatePassword :exec
UPDATE users
SET password_hash = $2,
    reset_token = NULL,
    reset_token_expires_at = NULL,
    updated_at = NOW()
WHERE id = $1;

-- name: SetPasswordResetToken :exec
UPDATE users
SET reset_token = $2,
    reset_token_expires_at = $3,
    updated_at = NOW()
WHERE id = $1;

-- name: GetUserByResetToken :one
SELECT *
FROM users
WHERE reset_token = $1
  AND reset_token_expires_at > NOW();

-- name: GetUserByID :one
SELECT *
FROM users
WHERE id = $1;

-- name: CreateDesktopLoginCode :one
INSERT INTO desktop_login_codes (code, user_id, state, expires_at)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: ConsumeDesktopLoginCode :one
UPDATE desktop_login_codes
SET used_at = NOW()
WHERE code = $1
  AND state = $2
  AND used_at IS NULL
  AND expires_at > NOW()
RETURNING *;

-- name: GetDesktopLoginCode :one
SELECT *
FROM desktop_login_codes
WHERE code = $1;

-- name: DeleteExpiredDesktopCodes :exec
DELETE FROM desktop_login_codes
WHERE expires_at < NOW()
   OR used_at IS NOT NULL;

-- 
-- Data Sync Queries
--

-- name: CreateBoard :one
INSERT INTO boards (id, user_id, name, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: CreateColumn :one
INSERT INTO columns (id, board_id, name, position, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: CreateCard :one
INSERT INTO cards (id, column_id, title, description, attachments, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: CreateTranscription :one
INSERT INTO transcriptions (id, board_id, transcription, recording_path, intent, assistant_response, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: GetUserBoards :many
SELECT * FROM boards
WHERE user_id = $1
ORDER BY created_at DESC;

-- name: GetBoardColumns :many
SELECT c.* FROM columns c
JOIN boards b ON b.id = c.board_id
WHERE c.board_id = $1 AND b.user_id = $2
ORDER BY position ASC;

-- name: GetColumnCards :many
SELECT * FROM cards
WHERE column_id = $1
ORDER BY created_at ASC;

-- name: GetBoardTranscriptions :many
SELECT * FROM transcriptions
WHERE board_id = $1
ORDER BY created_at DESC;

---- Operations -------

-- name: SyncUpsertBoard :exec
INSERT INTO boards (id, user_id, name, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name,
    updated_at = EXCLUDED.updated_at;

-- name: SyncDeleteBoard :exec
DELETE FROM boards
WHERE id = $1
  AND user_id = $2;

-- name: SyncUpsertColumn :exec
INSERT INTO columns (id, board_id, name, position, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (id) DO UPDATE SET
    board_id = EXCLUDED.board_id,
    name = EXCLUDED.name,
    position = EXCLUDED.position,
    updated_at = EXCLUDED.updated_at;

-- name: SyncDeleteColumn :exec
DELETE FROM columns c
USING boards b
WHERE c.id = $1
  AND b.id = c.board_id
  AND b.user_id = $2;

-- name: SyncUpsertCard :exec
INSERT INTO cards (id, column_id, title, description, attachments, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7)
ON CONFLICT (id) DO UPDATE SET
    column_id = EXCLUDED.column_id,
    title = EXCLUDED.title,
    description = EXCLUDED.description,
    attachments = EXCLUDED.attachments,
    updated_at = EXCLUDED.updated_at;

-- name: SyncDeleteCard :exec
DELETE FROM cards c
USING columns col
JOIN boards b ON b.id = col.board_id
WHERE c.id = $1
  AND c.column_id = col.id
  AND b.user_id = $2;

-- name: SyncUpdateCardColumn :exec
UPDATE cards c
SET column_id = $1,
    updated_at = $2
FROM columns col
JOIN boards b ON b.id = col.board_id
WHERE c.id = $3
  AND col.id = $1
  AND b.user_id = $4;

-- name: SyncPullColumns :many
SELECT c.id, c.board_id, c.name, c.position, c.created_at, c.updated_at
  FROM columns c
  JOIN boards b ON b.id = c.board_id
  WHERE b.user_id = $1;

--- TODO: created_at in operation table shouldnt be text, update it & update this function
-- name: GetAllOperations :many
SELECT o.*
FROM operations AS o
JOIN (
    SELECT inner_op.record_id, MAX(inner_op.created_at) AS max_created_at
    FROM operations AS inner_op
    WHERE inner_op.created_at > (
        SELECT COALESCE(
            to_char(to_timestamp(ss.last_synced_at), 'YYYY-MM-DD"T"HH24:MI:SS'),
            '1970-01-01T00:00:00'
        )
        FROM sync_state AS ss
        WHERE ss."table_name" = $1
          AND ss."user_id" = $4
    )
    AND inner_op."table_name" = $2
    GROUP BY inner_op.record_id
) AS latest
  ON o.record_id = latest.record_id
  AND o.created_at = latest.max_created_at
  AND o."table_name" = $3
JOIN columns AS c ON c.id = o.record_id
JOIN boards AS b ON b.id = c.board_id
WHERE b.user_id = $4
ORDER BY o.created_at ASC;


-- name: UpsertSyncState :exec
INSERT INTO sync_state (table_name, last_synced_at, last_synced_op_id, user_id)
VALUES ($1, $2, $3, $4)
ON CONFLICT(table_name, user_id)
DO UPDATE SET
  last_synced_at = EXCLUDED.last_synced_at,
  last_synced_op_id = EXCLUDED.last_synced_op_id;

-- name: GetSyncState :one
SELECT ss.*
FROM sync_state ss
WHERE ss.table_name = $1
  AND ss.user_id = $2
LIMIT 1;


-- name: UpdateSyncState :exec
UPDATE sync_state
SET last_synced_at = $1, last_synced_op_id = $2
WHERE table_name = $3
  AND user_id = $4;

-- name: InitCloud :exec
UPDATE "users"
SET cloud_initialized = true
WHERE id = $1;

-- name: GetCloudInitStatus :one
SELECT cloud_initialized 
FROM "users"
WHERE id = $1;

-- name: InitializeSyncStateForUser :exec
INSERT INTO sync_state (user_id, table_name, last_synced_at, last_synced_op_id)
VALUES 
    ($1, 'boards', EXTRACT(EPOCH FROM NOW())::BIGINT, NULL),
    ($1, 'columns', EXTRACT(EPOCH FROM NOW())::BIGINT, NULL),
    ($1, 'cards', EXTRACT(EPOCH FROM NOW())::BIGINT, NULL),
    ($1, 'transcriptions', EXTRACT(EPOCH FROM NOW())::BIGINT, NULL)
ON CONFLICT (user_id, table_name)
DO NOTHING;

-- name: CreateOperation :exec
INSERT INTO operations (
    id, table_name, record_id, operation_type, device_id, payload, created_at, updated_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
ON CONFLICT (id) DO UPDATE SET
    table_name = EXCLUDED.table_name,
    record_id = EXCLUDED.record_id,
    operation_type = EXCLUDED.operation_type,
    device_id = EXCLUDED.device_id,
    payload = EXCLUDED.payload,
    updated_at = EXCLUDED.updated_at;

-- name: SyncUpsertTranscription :exec
INSERT INTO transcriptions (id, board_id, transcription, recording_path, intent, assistant_response, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
ON CONFLICT (id) DO UPDATE SET
    board_id = EXCLUDED.board_id,
    transcription = EXCLUDED.transcription,
    recording_path = EXCLUDED.recording_path,
    intent = EXCLUDED.intent,
    assistant_response = EXCLUDED.assistant_response,
    updated_at = EXCLUDED.updated_at;

-- name: SyncDeleteTranscription :exec
DELETE FROM transcriptions t
USING boards b
WHERE t.id = $1
  AND t.board_id = b.id
  AND b.user_id = $2;

-- name: InsertBoardMember :exec
INSERT INTO board_members (board_id, user_id, role)
VALUES ($1, $2, $3)
ON CONFLICT (board_id, user_id) DO NOTHING;

-- name: RemoveBoardMember :exec
DELETE FROM board_members
WHERE board_id = $1 AND user_id = $2;

-- name: GetBoardMembers :many
SELECT u.*, bm.role, bm.joined_at
FROM board_members bm
JOIN users u ON bm.user_id = u.id
WHERE bm.board_id = $1
ORDER BY bm.joined_at ASC;

-- name: GetBoardsForUser :many
SELECT b.*
FROM boards b
JOIN board_members bm ON bm.board_id = b.id
WHERE bm.user_id = $1
ORDER BY b.created_at DESC;

-- name: IsUserMemberOfBoard :one
SELECT EXISTS (
  SELECT 1 FROM board_members
  WHERE board_id = $1 AND user_id = $2
) AS is_member;

-- name: ValidateBoardAccess :one
SELECT EXISTS (
  SELECT 1
  FROM board_members
  WHERE board_id = $1
    AND user_id = $2
) AS has_access;

-- name: GetColumnByID :one
SELECT * FROM columns
WHERE id = $1;


-- name: GetBoardByID :one
SELECT * FROM boards
WHERE id = $1;

-- name: EnsureBoardOwner :one
SELECT EXISTS (
  SELECT 1
  FROM board_members
  WHERE board_id = $1
    AND user_id = $2
    AND role = 'owner'
) AS is_owner;

-- name: ListBoards :many
SELECT * FROM boards
  WHERE user_id = $1
    ORDER BY created_at ASC
    LIMIT $2 OFFSET $3;

-- name: ListAllColumns :many
SELECT c.* FROM columns c
  JOIN boards b
    ON b.user_id = $1
ORDER BY c.created_at ASC;

-- name: ListBoardsCards :many
SELECT c.*
FROM cards c
JOIN columns col ON c.column_id = col.id
JOIN boards b ON col.board_id = b.id
WHERE b.user_id = $1 AND col.board_id = $2
ORDER BY c.created_at ASC;

-- name: ListBoardTranscriptions :many
SELECT t.* FROM transcriptions t
  JOIN boards b
    ON b.user_id = $1
WHERE b.id = $2
ORDER BY t.created_at DESC
LIMIT $3 OFFSET $4;

-- name: ListBoardMembers :many
SELECT bm.*, u.email
FROM board_members bm
JOIN users u ON u.id = bm.user_id
WHERE bm.board_id = $1
ORDER BY bm.joined_at DESC;

-- name: GetBoardMetadata :one
SELECT 
    b.id,
    b.name,
    b.created_at,
    b.updated_at,
    b.user_id,
    (SELECT COUNT(*) FROM columns c WHERE c.board_id = b.id) AS columns_count,
    (SELECT COUNT(*) FROM cards ca 
     JOIN columns col ON ca.column_id = col.id 
     WHERE col.board_id = b.id) AS cards_count,
    (SELECT COUNT(*) FROM transcriptions t WHERE t.board_id = b.id) AS transcriptions_count
FROM boards b
WHERE b.id = $1;

