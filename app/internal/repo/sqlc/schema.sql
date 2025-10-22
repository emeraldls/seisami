-- 1. Board Table
CREATE TABLE IF NOT EXISTS boards (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    created_at TEXT DEFAULT (datetime('now')),
    updated_at TEXT DEFAULT (datetime('now'))
);

-- 2. Column Table
CREATE TABLE IF NOT EXISTS "columns" (
    id TEXT PRIMARY KEY,
    board_id TEXT NOT NULL,
    name TEXT NOT NULL,
    position INTEGER NOT NULL,
    created_at TEXT DEFAULT (datetime('now')),
    updated_at TEXT DEFAULT (datetime('now')),
    FOREIGN KEY (board_id) REFERENCES boards(id) ON DELETE CASCADE
);

-- 3. Ticket Table
CREATE TABLE IF NOT EXISTS tickets (
    id TEXT PRIMARY KEY,
    column_id TEXT NOT NULL,
    title TEXT NOT NULL,
    description TEXT,
    assignee_id INTEGER,
    story_points INTEGER,
    pr_link TEXT,
    ticket_type TEXT NOT NULL,
    created_at TEXT DEFAULT (datetime('now')),
    updated_at TEXT DEFAULT (datetime('now')),
    FOREIGN KEY (column_id) REFERENCES columns(id) ON DELETE CASCADE
);


-- 4. Transcription Table
CREATE TABLE IF NOT EXISTS transcriptions (
    id TEXT PRIMARY KEY,
    board_id TEXT NOT NULL,
    transcription TEXT NOT NULL,
    recording_path TEXT,
    intent TEXT,
    assistant_response TEXT,
    created_at TEXT DEFAULT (datetime('now')),
    updated_at TEXT DEFAULT (datetime('now')),
    FOREIGN KEY (board_id) REFERENCES boards(id) ON DELETE CASCADE
);

-- 5. Settings Table
CREATE TABLE IF NOT EXISTS settings (
    id INTEGER PRIMARY KEY CHECK (id = 1), -- Single row table
    transcription_method TEXT NOT NULL DEFAULT 'cloud', -- 'cloud', 'local', 'custom'
    whisper_binary_path TEXT,
    whisper_model_path TEXT,
    openai_api_key TEXT,
    created_at TEXT DEFAULT (datetime('now')),
    updated_at TEXT DEFAULT (datetime('now'))
);

DROP TABLE tickets;

-- 6. Card Table
CREATE TABLE IF NOT EXISTS cards (
    id TEXT PRIMARY KEY,
    column_id TEXT NOT NULL,
    title TEXT NOT NULL,
    description TEXT,
    attachments TEXT,
    created_at TEXT DEFAULT (datetime('now')),
    updated_at TEXT DEFAULT (datetime('now')),
    FOREIGN KEY (column_id) REFERENCES columns(id) ON DELETE CASCADE
);

CREATE TRIGGER IF NOT EXISTS update_board_updated_at
AFTER UPDATE ON "boards"
FOR EACH ROW
BEGIN
  UPDATE "boards" SET updated_at = datetime('now') WHERE id = OLD.id;
END;

CREATE TRIGGER IF NOT EXISTS update_columns_updated_at
AFTER UPDATE ON "columns"
FOR EACH ROW
BEGIN
  UPDATE "columns" SET updated_at = datetime('now') WHERE id = OLD.id;
END;

CREATE TRIGGER IF NOT EXISTS update_cards_updated_at
AFTER UPDATE ON "cards"
FOR EACH ROW
BEGIN
  UPDATE "cards" SET updated_at = datetime('now') WHERE id = OLD.id;
END;

CREATE TRIGGER IF NOT EXISTS update_settings_updated_at
AFTER UPDATE ON "settings"
FOR EACH ROW
BEGIN
  UPDATE "settings" SET updated_at = datetime('now') WHERE id = OLD.id;
END;