package db

import (
	"os"

	"github.com/jmoiron/sqlx"
	"pawked.com/sendyourpicks/internal/logger"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// sets up the connection to supabase
func NewPostgres() (*sqlx.DB, error) {
	logger.Info("Connecting to database...")

	dsn := os.Getenv("DATABASE_URL")

	//logger.Debug("Connection string", "dsn", dsn)

	db, err := sqlx.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	logger.Info("Database connection successful")
	return db, nil
}
