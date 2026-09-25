PRAGMA journal_mode = WAL;
PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS habit_schemas (
    habit_id TEXT NOT NULL,
    version INTEGER NOT NULL DEFAULT 1,
    name TEXT NOT NULL,
    description TEXT DEFAULT '',
    json_schema JSON NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (habit_id, version),
    CONSTRAINT valid_json_schema CHECK (json_valid(json_schema))
);

CREATE INDEX IF NOT EXISTS idx_schemas_habit_version
ON habit_schemas(habit_id, version DESC);

CREATE TABLE IF NOT EXISTS habit_entries (
    id TEXT PRIMARY KEY,
    habit_id TEXT NOT NULL,
    schema_version INTEGER NOT NULL,
    data JSON NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (habit_id, schema_version) 
        REFERENCES habit_schemas(habit_id, version) 
        ON DELETE CASCADE,
    CONSTRAINT valid_entry_data CHECK (json_valid(data))
);

-- Faster lookup on latest additions
CREATE INDEX IF NOT EXISTS idx_entries_habit_created 
ON habit_entries(habit_id, created_at DESC);

CREATE TRIGGER IF NOT EXISTS update_habit_entries_timestamp 
AFTER UPDATE ON habit_entries
BEGIN
    UPDATE habit_entries 
    SET updated_at = CURRENT_TIMESTAMP 
    WHERE id = NEW.id;
END;
