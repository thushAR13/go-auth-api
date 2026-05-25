package db

import (
	"database/sql"
	"fmt"
	"go-auth-api/config"
	"log/slog"
	"os"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func Init(cfg *config.Config) {
	// connStr := "postgres://postgres:1111@localhost:5432/goauthapi?sslmode=disable"
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName)

	var err error

	DB, err = sql.Open("postgres", connStr)

	if err != nil {
		slog.Error("Failed to open database connection",
			"error", err)
		os.Exit(1)
	}

	if err = DB.Ping(); err != nil {
		slog.Error("Failed to ping database",
			"error", err)
	}
	slog.Info("Database connected succesfully")
}
