package db

import (
	"database/sql"
	"fmt"
	"go-auth-api/config"
	"log/slog"
	"os"
	"time"

	_ "github.com/lib/pq"
)

type Store struct {
	DB *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{DB: db}
}
func Init(cfg *config.Config) *sql.DB {
	// connStr := "postgres://postgres:1111@localhost:5432/goauthapi?sslmode=disable"
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName)

	var err error
	var db *sql.DB
	db, err = sql.Open("postgres", connStr)
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(1 * time.Hour)
	db.SetConnMaxIdleTime(30 * time.Minute)

	if err != nil {
		slog.Error("Failed to open database connection",
			"error", err)
		os.Exit(1)
	}

	if err = db.Ping(); err != nil {
		slog.Error("Failed to ping database",
			"error", err)
		os.Exit(1)
	}
	slog.Info("Database connected succesfully")
	return db
}
