-- 1. Habit Schemas Definition Table
CREATE TABLE IF NOT EXISTS habit_schemas (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    schema_definition JSON NOT NULL, -- JSON Schema rules (types, requirements)
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT valid_schema CHECK (json_valid(schema_definition))
);

-- 2. Habit Entries Log Table
CREATE TABLE IF NOT EXISTS habit_entries (
    id TEXT PRIMARY KEY,
    schema_id TEXT NOT NULL,
    data JSON NOT NULL,               -- The actual submitted entry payload
    logged_at DATETIME NOT NULL,     -- Timestamp when the habit occurred
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (schema_id) REFERENCES habit_schemas(id) ON DELETE CASCADE,
    CONSTRAINT valid_data CHECK (json_valid(data))
);

-- Index for fast lookup when browsing habit logs chronologically
CREATE INDEX IF NOT EXISTS idx_entries_schema_logged 
ON habit_entries(schema_id, logged_at DESC);
