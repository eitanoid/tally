package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"embed"

	"github.com/eitanoid/habit-tracker/internal/entries"
	"github.com/eitanoid/habit-tracker/internal/schemas"

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

func (c *SqliteClient) InsertSchema(ctx context.Context, schema *schemas.HabitSchema) error {
	query := `
		INSERT INTO habit_schemas (habit_id, version, name, description, json_schema)
		VALUES (?, ?, ?, ?, ?)
		RETURNING created_at;
	`
	err := c.db.QueryRowContext(ctx, query,
		schema.HabitID,
		schema.Version,
		schema.Name,
		schema.Description,
		schema.JSONSchemaRaw,
	).Scan(&schema.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to insert habit schema: %w", err)
	}
	return nil
}

func (c *SqliteClient) GetLatestSchemaByID(ctx context.Context, habitID string) (*schemas.HabitSchema, error) {
	query := `
		SELECT habit_id, version, name, description, json_schema, created_at
		FROM habit_schemas
		WHERE habit_id = ?
		ORDER BY version DESC
		LIMIT 1;
	`
	var s schemas.HabitSchema
	err := c.db.QueryRowContext(ctx, query, habitID).Scan(
		&s.HabitID,
		&s.Version,
		&s.Name,
		&s.Description,
		&s.JSONSchemaRaw,
		&s.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("schema not found for habit_id %s: %w", habitID, err)
		}
		return nil, fmt.Errorf("failed to query latest schema: %w", err)
	}
	return &s, nil
}

func (c *SqliteClient) GetSchemaByRef(ctx context.Context, ref schemas.SchemaRef) (*schemas.HabitSchema, error) {
	query := `
		SELECT habit_id, version, name, description, json_schema, created_at
		FROM habit_schemas
		WHERE habit_id = ? AND version = ?;
	`
	var s schemas.HabitSchema
	err := c.db.QueryRowContext(ctx, query, ref.HabitID, ref.Version).Scan(
		&s.HabitID,
		&s.Version,
		&s.Name,
		&s.Description,
		&s.JSONSchemaRaw,
		&s.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("schema not found for ref (%s v%d): %w", ref.HabitID, ref.Version, err)
		}
		return nil, fmt.Errorf("failed to query schema by ref: %w", err)
	}
	return &s, nil
}

func (c *SqliteClient) GetAllLatestSchemas(ctx context.Context) ([]schemas.HabitSchema, error) {
	query := `
		SELECT habit_id, version, name, description, json_schema, created_at
		FROM habit_schemas
		WHERE (habit_id, version) IN (
			SELECT habit_id, MAX(version)
			FROM habit_schemas
			GROUP BY habit_id
		)
		ORDER BY name ASC;
	`
	rows, err := c.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query all latest schemas: %w", err)
	}
	defer rows.Close()

	var result []schemas.HabitSchema
	for rows.Next() {
		var s schemas.HabitSchema
		if err := rows.Scan(&s.HabitID, &s.Version, &s.Name, &s.Description, &s.JSONSchemaRaw, &s.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan schema row: %w", err)
		}
		result = append(result, s)
	}
	return result, nil
}

// -----------------------------------------------------------------------------
// Entry Operations
// -----------------------------------------------------------------------------

func (c *SqliteClient) InsertEntry(ctx context.Context, entry *entries.HabitEntry) error {
	query := `
		INSERT INTO habit_entries (id, habit_id, schema_version, data)
		VALUES (?, ?, ?, ?)
		RETURNING created_at, updated_at;
	`
	err := c.db.QueryRowContext(ctx, query,
		entry.ID,
		entry.HabitID,
		entry.SchemaVersion,
		entry.Data,
	).Scan(&entry.CreatedAt, &entry.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to insert habit entry: %w", err)
	}
	return nil
}

func (c *SqliteClient) GetEntriesByHabitID(ctx context.Context, habitID string, limit int) ([]entries.HabitEntry, error) {
	if limit <= 0 {
		limit = 50
	}

	query := `
		SELECT id, habit_id, schema_version, data, created_at, updated_at
		FROM habit_entries
		WHERE habit_id = ?
		ORDER BY created_at DESC
		LIMIT ?;
	`
	rows, err := c.db.QueryContext(ctx, query, habitID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query entries for habit %s: %w", habitID, err)
	}
	defer rows.Close()

	var result []entries.HabitEntry
	for rows.Next() {
		var e entries.HabitEntry
		if err := rows.Scan(&e.ID, &e.HabitID, &e.SchemaVersion, &e.Data, &e.CreatedAt, &e.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan entry row: %w", err)
		}
		result = append(result, e)
	}
	return result, nil
}
