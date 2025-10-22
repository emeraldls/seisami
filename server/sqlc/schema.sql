CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY,
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    reset_token TEXT,
    reset_token_expires_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS desktop_login_codes (
    code TEXT PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    state TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL,
    used_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS desktop_login_codes_expires_idx
    ON desktop_login_codes (expires_at);

CREATE INDEX IF NOT EXISTS desktop_login_codes_state_idx
    ON desktop_login_codes (state);
