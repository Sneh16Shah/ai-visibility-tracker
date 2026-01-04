package db

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/Sneh16Shah/ai-visibility-tracker/config"
	_ "github.com/go-sql-driver/mysql"
)

var DB *sql.DB

// Connect establishes a connection to the MySQL database
func Connect(cfg *config.Config) error {
	dsn := cfg.GetDSN()

	var err error
	DB, err = sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	DB.SetMaxOpenConns(25)
	DB.SetMaxIdleConns(5)
	DB.SetConnMaxLifetime(5 * time.Minute)

	// Test the connection
	if err = DB.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("✅ Database connected successfully")
	return nil
}

// Close closes the database connection
func Close() {
	if DB != nil {
		DB.Close()
		log.Println("Database connection closed")
	}
}

// GetDB returns the database instance
func GetDB() *sql.DB {
	return DB
}
