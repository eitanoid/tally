package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"embed"

	"github.com/eitanoid/tally/internal/entries"
	"github.com/eitanoid/tally/internal/schemas"

	_ "github.com/ncruces/go-sqlite3/driver"
)

type SqliteClient struct {
	db *sql.DB
}

var _ Repository = (*SqliteClient)(nil)

//go:embed migrations/*.sql
var migrationFS embed.FS

func NewSQLiteClient(dbPath string) (*SqliteClient, error) {
	db, err := open(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create SQLite database: %w", err)
	}
	return &SqliteClient{
		db: db,
	}, nil
}

func (c *SqliteClient) Close() error {
	return c.db.Close()
}

// Open initializes the SQLite database connection, applies Pragmas, and executes embedded migrations.
func open(dbPath string) (*sql.DB, error) {
	dsn := fmt.Sprintf("file:%s?_journal_mode=WAL&_busy_timeout=5000&_foreign_keys=ON", dbPath)

	database, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite database: %w", err)
	}

	// Embedded SQLite connection pool tuning
	database.SetMaxOpenConns(1)
	database.SetMaxIdleConns(1)
	database.SetConnMaxLifetime(time.Hour)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := database.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	if err := runMigrations(ctx, database); err != nil {
		return nil, fmt.Errorf("failed to execute migrations: %w", err)
	}

	return database, nil
}

func runMigrations(ctx context.Context, db *sql.DB) error {
	migrationScript, err := migrationFS.ReadFile("migrations/000001_init_schema.sql")
	if err != nil {
		return fmt.Errorf("failed to read embedded migration file: %w", err)
	}

	if _, err := db.ExecContext(ctx, string(migrationScript)); err != nil {
		return fmt.Errorf("failed to execute migration script: %w", err)
	}

	return nil
}

// --

func (c *SqliteClient) InsertSchema(ctx context.Context, schema *schemas.TallySchema) error {
	query := `
		INSERT INTO tally_schemas (tally_id, version, name, description, json_schema)
		VALUES (?, ?, ?, ?, ?)
		RETURNING created_at;
	`
	err := c.db.QueryRowContext(ctx, query,
		schema.TallyID,
		schema.Version,
		schema.Name,
		schema.Description,
		schema.JSONSchemaRaw,
	).Scan(&schema.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to insert tally schema: %w", err)
	}
	return nil
}

func (c *SqliteClient) GetLatestSchemaByID(ctx context.Context, tallyID string) (*schemas.TallySchema, error) {
	query := `
		SELECT tally_id, version, name, description, json_schema, created_at
		FROM tally_schemas
		WHERE tally_id = ?
		ORDER BY version DESC
		LIMIT 1;
	`
	var s schemas.TallySchema
	err := c.db.QueryRowContext(ctx, query, tallyID).Scan(
		&s.TallyID,
		&s.Version,
		&s.Name,
		&s.Description,
		&s.JSONSchemaRaw,
		&s.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("schema not found for tally_id %s: %w", tallyID, err)
		}
		return nil, fmt.Errorf("failed to query latest schema: %w", err)
	}
	return &s, nil
}

func (c *SqliteClient) GetSchemaByRef(ctx context.Context, ref schemas.SchemaRef) (*schemas.TallySchema, error) {
	query := `
		SELECT tally_id, version, name, description, json_schema, created_at
		FROM tally_schemas
		WHERE tally_id = ? AND version = ?;
	`
	var s schemas.TallySchema
	err := c.db.QueryRowContext(ctx, query, ref.TallyID, ref.Version).Scan(
		&s.TallyID,
		&s.Version,
		&s.Name,
		&s.Description,
		&s.JSONSchemaRaw,
		&s.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("schema not found for ref (%s v%d): %w", ref.TallyID, ref.Version, err)
		}
		return nil, fmt.Errorf("failed to query schema by ref: %w", err)
	}
	return &s, nil
}

func (c *SqliteClient) GetAllLatestSchemas(ctx context.Context) ([]schemas.TallySchema, error) {
	query := `
		SELECT tally_id, version, name, description, json_schema, created_at
		FROM tally_schemas
		WHERE (tally_id, version) IN (
			SELECT tally_id, MAX(version)
			FROM tally_schemas
			GROUP BY tally_id
		)
		ORDER BY name ASC;
	`
	rows, err := c.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query all latest schemas: %w", err)
	}
	defer rows.Close()

	var result []schemas.TallySchema
	for rows.Next() {
		var s schemas.TallySchema
		if err := rows.Scan(&s.TallyID, &s.Version, &s.Name, &s.Description, &s.JSONSchemaRaw, &s.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan schema row: %w", err)
		}
		result = append(result, s)
	}
	return result, nil
}

// -----------------------------------------------------------------------------
// Entry Operations
// -----------------------------------------------------------------------------

func (c *SqliteClient) InsertEntry(ctx context.Context, entry *entries.TallyEntry) error {
	query := `
		INSERT INTO tally_entries (id, tally_id, schema_version, data)
		VALUES (?, ?, ?, ?)
		RETURNING created_at, updated_at;
	`
	err := c.db.QueryRowContext(ctx, query,
		entry.ID,
		entry.TallyID,
		entry.SchemaVersion,
		entry.Data,
	).Scan(&entry.CreatedAt, &entry.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to insert tally entry: %w", err)
	}
	return nil
}

// GetEntriesByTallyID retrieves a paginated slice of entries alongside the total entry count.
func (c *SqliteClient) GetEntriesByTallyID(ctx context.Context, filter EntryFilter) ([]entries.TallyEntry, int, error) {

	var totalCount int
	countQuery := `SELECT COUNT(*) FROM entries WHERE tally_id = ?`
	if err := c.db.QueryRowContext(ctx, countQuery, filter.TallyID).Scan(&totalCount); err != nil {
		return nil, 0, fmt.Errorf("failed to count entries: %w", err)
	}

	// Clamp limit
	if filter.Limit <= 0 {
		filter.Limit = 50
	}
	if filter.Limit > 200 {
		filter.Limit = 200
	}

	query := `
		SELECT id, tally_id, schema_version, data, created_at, updated_at
		FROM tally_entries
		WHERE tally_id = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?;
	`
	rows, err := c.db.QueryContext(ctx, query, filter.TallyID, filter.Limit, filter.Offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query entries for tally %s: %w", filter.TallyID, err)
	}
	defer rows.Close()

	var result []entries.TallyEntry
	for rows.Next() {
		var e entries.TallyEntry
		if err := rows.Scan(&e.ID, &e.TallyID, &e.SchemaVersion, &e.Data, &e.CreatedAt, &e.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("failed to scan entry row: %w", err)
		}
		result = append(result, e)
	}
	return result, totalCount, nil
}
