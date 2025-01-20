package postgres

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMigrations(t *testing.T) {
	helper := setupTestDB(t)
	defer helper.cleanup(t)

	store := helper.createTestStore(t)
	defer store.Close()

	t.Run("InitialMigration", func(t *testing.T) {
		// Apply migrations
		err := store.Migrate(helper.ctx)
		require.NoError(t, err)

		// Check if migrations were applied
		applied, err := store.GetAppliedMigrations(helper.ctx)
		require.NoError(t, err)
		assert.NotEmpty(t, applied)

		// Verify tables were created
		tables := []string{"users", "sessions", "messages", "user_sessions"}
		for _, table := range tables {
			var exists bool
			err := helper.db.QueryRow(helper.ctx, `
				SELECT EXISTS (
					SELECT FROM information_schema.tables 
					WHERE table_schema = 'public' 
					AND table_name = $1
				)
			`, table).Scan(&exists)
			require.NoError(t, err)
			assert.True(t, exists, "Table %s should exist", table)
		}
	})

	t.Run("ReapplyMigrations", func(t *testing.T) {
		// Try to apply migrations again
		err := store.Migrate(helper.ctx)
		require.NoError(t, err)

		// Get applied migrations
		applied, err := store.GetAppliedMigrations(helper.ctx)
		require.NoError(t, err)

		// Count should remain the same
		migrations, err := store.migration.LoadMigrations()
		require.NoError(t, err)
		assert.Equal(t, len(migrations), len(applied))
	})

	t.Run("RollbackMigration", func(t *testing.T) {
		// Get initial count of applied migrations
		initialApplied, err := store.GetAppliedMigrations(helper.ctx)
		require.NoError(t, err)
		initialCount := len(initialApplied)

		// Rollback last migration
		err = store.Rollback(helper.ctx)
		require.NoError(t, err)

		// Check if one migration was removed
		currentApplied, err := store.GetAppliedMigrations(helper.ctx)
		require.NoError(t, err)
		assert.Equal(t, initialCount-1, len(currentApplied))

		// Reapply migration
		err = store.Migrate(helper.ctx)
		require.NoError(t, err)

		// Check if migration was reapplied
		finalApplied, err := store.GetAppliedMigrations(helper.ctx)
		require.NoError(t, err)
		assert.Equal(t, initialCount, len(finalApplied))
	})

	t.Run("MigrationOrder", func(t *testing.T) {
		// Get applied migrations
		applied, err := store.GetAppliedMigrations(helper.ctx)
		require.NoError(t, err)

		// Check if migrations were applied in order
		var lastVersion int
		var lastAppliedAt time.Time
		for version, migration := range applied {
			assert.Greater(t, version, lastVersion)
			if !lastAppliedAt.IsZero() {
				assert.True(t, migration.AppliedAt.After(lastAppliedAt) ||
					migration.AppliedAt.Equal(lastAppliedAt))
			}
			lastVersion = version
			lastAppliedAt = migration.AppliedAt
		}
	})

	t.Run("LoadMigrations", func(t *testing.T) {
		migrations, err := store.migration.LoadMigrations()
		require.NoError(t, err)
		assert.NotEmpty(t, migrations)

		// Check migration format
		for _, migration := range migrations {
			assert.Greater(t, migration.Version, 0)
			assert.NotEmpty(t, migration.Name)
			assert.NotEmpty(t, migration.UpSQL)
			assert.NotEmpty(t, migration.DownSQL)
			assert.True(t, migration.AppliedAt.IsZero())
		}

		// Check migrations are sorted
		for i := 1; i < len(migrations); i++ {
			assert.Greater(t, migrations[i].Version, migrations[i-1].Version)
		}
	})
}

func TestMigrationManager_Initialize(t *testing.T) {
	helper := setupTestDB(t)
	defer helper.cleanup(t)

	store := helper.createTestStore(t)
	defer store.Close()

	t.Run("CreateMigrationsTable", func(t *testing.T) {
		// Initialize migrations table
		err := store.migration.Initialize(helper.ctx)
		require.NoError(t, err)

		// Check if table exists
		var exists bool
		err = helper.db.QueryRow(helper.ctx, `
			SELECT EXISTS (
				SELECT FROM information_schema.tables 
				WHERE table_schema = 'public' 
				AND table_name = 'schema_migrations'
			)
		`).Scan(&exists)
		require.NoError(t, err)
		assert.True(t, exists)

		// Check table structure
		columns := []struct {
			name     string
			dataType string
			nullable string
		}{
			{"version", "integer", "NO"},
			{"name", "text", "NO"},
			{"applied_at", "timestamp with time zone", "NO"},
		}

		for _, col := range columns {
			var colName, dataType, nullable string
			err := helper.db.QueryRow(helper.ctx, `
				SELECT column_name, data_type, is_nullable
				FROM information_schema.columns
				WHERE table_schema = 'public'
				AND table_name = 'schema_migrations'
				AND column_name = $1
			`, col.name).Scan(&colName, &dataType, &nullable)
			require.NoError(t, err)
			assert.Equal(t, col.name, colName)
			assert.Equal(t, col.dataType, dataType)
			assert.Equal(t, col.nullable, nullable)
		}
	})
}
