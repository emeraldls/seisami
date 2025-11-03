CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY,
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    reset_token TEXT,
    reset_token_expires_at TIMESTAMPTZ,
    cloud_initialized BOOLEAN
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

-- Cloud data sync tables
CREATE TABLE IF NOT EXISTS boards (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS boards_user_id_idx ON boards(user_id);

CREATE TABLE IF NOT EXISTS columns (
    id TEXT PRIMARY KEY,
    board_id UUID NOT NULL REFERENCES boards(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    position INTEGER NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS columns_board_id_idx ON columns(board_id);

CREATE TABLE IF NOT EXISTS cards (
    id TEXT PRIMARY KEY,
    column_id TEXT NOT NULL REFERENCES columns(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    description TEXT,
    attachments TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS cards_column_id_idx ON cards(column_id);

CREATE TABLE IF NOT EXISTS transcriptions (
    id TEXT PRIMARY KEY,
    board_id UUID NOT NULL REFERENCES boards(id) ON DELETE CASCADE,
    transcription TEXT NOT NULL,
    recording_path TEXT,
    intent TEXT,
    assistant_response TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS transcriptions_board_id_idx ON transcriptions(board_id);

CREATE TABLE IF NOT EXISTS operations (
  id TEXT PRIMARY KEY,
  "table_name" TEXT NOT NULL,
  record_id TEXT NOT NULL,
  operation_type TEXT NOT NULL,
  device_id TEXT, 
  payload TEXT NOT NULL,
  created_at TEXT DEFAULT CURRENT_TIMESTAMP,
  updated_at TEXT DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS sync_state (
  user_id UUID NOT NULL REFERENCES users(id),
  "table_name" TEXT NOT NULL,
  last_synced_at BIGINT NOT NULL,
  last_synced_op_id TEXT,
  PRIMARY KEY (user_id, table_name)
);

CREATE TABLE IF NOT EXISTS board_members (
  board_id UUID NOT NULL REFERENCES boards(id) ON DELETE CASCADE,
  user_id  UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  role     TEXT DEFAULT 'member', 
  joined_at TIMESTAMPTZ DEFAULT NOW(),
  PRIMARY KEY (board_id, user_id)
);

CREATE TABLE IF NOT EXISTS app_versions (
  id SERIAL PRIMARY KEY,
  version TEXT UNIQUE NOT NULL,
  url TEXT NOT NULL,
  notes TEXT,
  sha256 TEXT,
  created_at TIMESTAMPTZ DEFAULT now()
);
