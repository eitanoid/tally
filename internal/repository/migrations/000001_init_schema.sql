PRAGMA journal_mode = WAL;
PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS tally_schemas (
    tally_id TEXT NOT NULL,
    version INTEGER NOT NULL DEFAULT 1,
    name TEXT NOT NULL,
    description TEXT DEFAULT '',
    json_schema JSON NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (tally_id, version),
    CONSTRAINT valid_json_schema CHECK (json_valid(json_schema))
);

CREATE INDEX IF NOT EXISTS idx_schemas_tally_version
ON tally_schemas(tally_id, version DESC);

CREATE TABLE IF NOT EXISTS tally_entries (
    id TEXT PRIMARY KEY,
    tally_id TEXT NOT NULL,
    schema_version INTEGER NOT NULL,
    data JSON NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (tally_id, schema_version) 
        REFERENCES tally_schemas(tally_id, version) 
        ON DELETE CASCADE,
    CONSTRAINT valid_entry_data CHECK (json_valid(data))
);

-- Faster lookup on latest additions
CREATE INDEX IF NOT EXISTS idx_entries_tally_created 
ON tally_entries(tally_id, created_at DESC);

CREATE TRIGGER IF NOT EXISTS update_tally_entries_timestamp 
AFTER UPDATE ON tally_entries
BEGIN
    UPDATE tally_entries 
    SET updated_at = CURRENT_TIMESTAMP 
    WHERE id = NEW.id;
END;
