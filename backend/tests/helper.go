package tests

import (
	"chat-room/config"
	"chat-room/database"
	"testing"

	"gorm.io/gorm"
)

func SetupTestDB(t *testing.T) *gorm.DB {
	// Use test database
	cfg := &config.Config{
		DBHost:     "localhost",
		DBPort:     "5432",
		DBUser:     "postgres",
		DBPassword: "your_password",
		DBName:     "chatroom_test",
	}

	db, err := database.InitDB(cfg)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Clean up database after test
	t.Cleanup(func() {
		sqlDB, err := db.DB()
		if err != nil {
			t.Errorf("Failed to get database instance: %v", err)
			return
		}
		sqlDB.Close()
	})

	return db
}

func CleanupTestDB(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("DELETE FROM messages").Error; err != nil {
			return err
		}
		if err := tx.Exec("DELETE FROM user_sessions").Error; err != nil {
			return err
		}
		if err := tx.Exec("DELETE FROM sessions").Error; err != nil {
			return err
		}
		if err := tx.Exec("DELETE FROM users").Error; err != nil {
			return err
		}
		return nil
	})
}
