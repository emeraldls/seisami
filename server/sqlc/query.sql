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
