package testing

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// TestDB provides a test database for integration tests
type TestDB struct {
	DB     *gorm.DB
	DSN    string
	dbName string
}

// NewTestDB creates a new test database
func NewTestDB(t *testing.T) *TestDB {
	// Get connection details from environment or use defaults
	host := getEnv("TEST_DB_HOST", "localhost")
	port := getEnv("TEST_DB_PORT", "5432")
	user := getEnv("TEST_DB_USER", "postgres")
	password := getEnv("TEST_DB_PASSWORD", "postgres")
	sslmode := getEnv("TEST_DB_SSLMODE", "disable")

	// Create a unique database name for this test
	dbName := fmt.Sprintf("test_%s", uuid.New().String()[:8])

	// Connect to the default database first
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s sslmode=%s dbname=postgres",
		host, port, user, password, sslmode)

	mainDB, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("Failed to connect to main database: %v", err)
	}
	defer func() {
		if err := mainDB.Close(); err != nil {
			fmt.Println("error closing mainDB:", err)
		}
	}()

	// Create the test database
	_, err = mainDB.Exec(fmt.Sprintf("CREATE DATABASE %s", dbName))
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Connect to the test database
	testDSN := fmt.Sprintf("host=%s port=%s user=%s password=%s sslmode=%s dbname=%s",
		host, port, user, password, sslmode, dbName)

	// Configure GORM
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	}

	gormDB, err := gorm.Open(postgres.Open(testDSN), gormConfig)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	return &TestDB{
		DB:     gormDB,
		DSN:    testDSN,
		dbName: dbName,
	}
}

// Close cleans up the test database
func (tdb *TestDB) Close(t *testing.T) {
	// Close the test database connection
	sqlDB, err := tdb.DB.DB()
	if err != nil {
		t.Errorf("Failed to get underlying database connection: %v", err)
		return
	}
	if err := sqlDB.Close(); err != nil {
		fmt.Println("error closing sqlDB:", err)
	}

	// Create a connection to drop the test database
	host := getEnv("TEST_DB_HOST", "localhost")
	port := getEnv("TEST_DB_PORT", "5432")
	user := getEnv("TEST_DB_USER", "postgres")
	password := getEnv("TEST_DB_PASSWORD", "postgres")
	sslmode := getEnv("TEST_DB_SSLMODE", "disable")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s sslmode=%s dbname=postgres",
		host, port, user, password, sslmode)

	mainDB, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Errorf("Failed to connect to main database for cleanup: %v", err)
		return
	}
	defer func() {
		if err := mainDB.Close(); err != nil {
			fmt.Println("error closing mainDB:", err)
		}
	}()

	// Close any existing connections
	_, err = mainDB.Exec(fmt.Sprintf("SELECT pg_terminate_backend(pg_stat_activity.pid) FROM pg_stat_activity WHERE pg_stat_activity.datname = '%s' AND pid <> pg_backend_pid()", tdb.dbName))
	if err != nil {
		t.Errorf("Failed to terminate connections: %v", err)
	}

	// Drop the test database
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = mainDB.ExecContext(ctx, fmt.Sprintf("DROP DATABASE IF EXISTS %s", tdb.dbName))
	if err != nil {
		t.Errorf("Failed to drop test database: %v", err)
	}
}

// getEnv gets an environment variable or returns the default value
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
