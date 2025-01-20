package postgres

import (
	"context"
	"embed"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

// migrationManager handles database migrations
type migrationManager struct {
	db      dbConn
	queries embed.FS
}

// migration represents a single database migration
type migration struct {
	Version   int
	Name      string
	UpSQL     string
	DownSQL   string
	AppliedAt time.Time
}

// newMigrationManager creates a new migration manager
func newMigrationManager(db dbConn, queries embed.FS) *migrationManager {
	return &migrationManager{
		db:      db,
		queries: queries,
	}
}

// Initialize creates the migrations table if it doesn't exist
func (m *migrationManager) Initialize(ctx context.Context) error {
	createTableSQL := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version     INTEGER PRIMARY KEY,
			name        TEXT NOT NULL,
			applied_at  TIMESTAMP WITH TIME ZONE NOT NULL
		);
	`
	_, err := m.db.Exec(ctx, createTableSQL)
	if err != nil {
		return fmt.Errorf("creating migrations table: %w", err)
	}
	return nil
}

// LoadMigrations loads all migrations from the embedded filesystem
func (m *migrationManager) LoadMigrations() ([]migration, error) {
	entries, err := m.queries.ReadDir("migrations")
	if err != nil {
		return nil, fmt.Errorf("reading migrations directory: %w", err)
	}

	var migrations []migration
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".sql") {
			content, err := m.queries.ReadFile("migrations/" + entry.Name())
			if err != nil {
				return nil, fmt.Errorf("reading migration file %s: %w", entry.Name(), err)
			}

			parts := strings.Split(strings.TrimSuffix(entry.Name(), ".sql"), "_")
			if len(parts) < 2 {
				return nil, fmt.Errorf("invalid migration filename: %s", entry.Name())
			}

			version := 0
			_, err = fmt.Sscanf(parts[0], "%d", &version)
			if err != nil {
				return nil, fmt.Errorf("parsing migration version from %s: %w", entry.Name(), err)
			}

			name := strings.Join(parts[1:], "_")
			sections := strings.Split(string(content), "-- Down")
			if len(sections) != 2 {
				return nil, fmt.Errorf("invalid migration format in %s", entry.Name())
			}

			upSQL := strings.TrimSpace(sections[0])
			downSQL := strings.TrimSpace(sections[1])

			migrations = append(migrations, migration{
				Version: version,
				Name:    name,
				UpSQL:   upSQL,
				DownSQL: downSQL,
			})
		}
	}

	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	return migrations, nil
}

// GetAppliedMigrations returns all migrations that have been applied
func (m *migrationManager) GetAppliedMigrations(ctx context.Context) (map[int]migration, error) {
	rows, err := m.db.Query(ctx, `
		SELECT version, name, applied_at
		FROM schema_migrations
		ORDER BY version ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("querying applied migrations: %w", err)
	}
	defer rows.Close()

	applied := make(map[int]migration)
	for rows.Next() {
		var mig migration
		err := rows.Scan(&mig.Version, &mig.Name, &mig.AppliedAt)
		if err != nil {
			return nil, fmt.Errorf("scanning migration row: %w", err)
		}
		applied[mig.Version] = mig
	}

	return applied, rows.Err()
}

// Migrate applies all pending migrations
func (m *migrationManager) Migrate(ctx context.Context) error {
	if err := m.Initialize(ctx); err != nil {
		return err
	}

	migrations, err := m.LoadMigrations()
	if err != nil {
		return err
	}

	applied, err := m.GetAppliedMigrations(ctx)
	if err != nil {
		return err
	}

	tx, ok := m.db.(pgx.Tx)
	if !ok {
		return fmt.Errorf("database connection is not a transaction")
	}

	for _, migration := range migrations {
		if _, exists := applied[migration.Version]; !exists {
			// Apply migration
			if _, err := tx.Exec(ctx, migration.UpSQL); err != nil {
				return fmt.Errorf("applying migration %d: %w", migration.Version, err)
			}

			// Record migration
			if _, err := tx.Exec(ctx, `
				INSERT INTO schema_migrations (version, name, applied_at)
				VALUES ($1, $2, $3)
			`, migration.Version, migration.Name, time.Now().UTC()); err != nil {
				return fmt.Errorf("recording migration %d: %w", migration.Version, err)
			}
		}
	}

	return nil
}

// Rollback reverts the last applied migration
func (m *migrationManager) Rollback(ctx context.Context) error {
	applied, err := m.GetAppliedMigrations(ctx)
	if err != nil {
		return err
	}

	if len(applied) == 0 {
		return nil
	}

	migrations, err := m.LoadMigrations()
	if err != nil {
		return err
	}

	var lastMigration migration
	var lastVersion int
	for version := range applied {
		if version > lastVersion {
			lastVersion = version
			for _, m := range migrations {
				if m.Version == version {
					lastMigration = m
					break
				}
			}
		}
	}

	tx, ok := m.db.(pgx.Tx)
	if !ok {
		return fmt.Errorf("database connection is not a transaction")
	}

	// Apply rollback
	if _, err := tx.Exec(ctx, lastMigration.DownSQL); err != nil {
		return fmt.Errorf("rolling back migration %d: %w", lastMigration.Version, err)
	}

	// Remove migration record
	if _, err := tx.Exec(ctx, `
		DELETE FROM schema_migrations
		WHERE version = $1
	`, lastMigration.Version); err != nil {
		return fmt.Errorf("removing migration record %d: %w", lastMigration.Version, err)
	}

	return nil
}
