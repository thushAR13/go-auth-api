package main

import (
	"encoding/json"
	"go-auth-api/config"
	"go-auth-api/db"
	"go-auth-api/handlers"
	"go-auth-api/middleware"
	"go-auth-api/services"
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

func main() {

	if err := godotenv.Load(); err != nil {
		slog.Warn("No .env file found, reading from environment")
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	slog.SetDefault(logger)

	slog.Info("Logger initialized")

	cfg, err := config.Load()
	if err != nil {
		slog.Error("Configuration error",
			"error", err)
		os.Exit(1)
	}
	db.Init(cfg)
	services.Init(cfg)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Welcome to go auth api!!\n"))
	})

	mux.Handle("POST /api/register", middleware.RateLmitMiddleWare(
		http.HandlerFunc(handlers.Register),
	))
	mux.Handle("POST /api/login", middleware.RateLmitMiddleWare(
		http.HandlerFunc(handlers.Login),
	))

	mux.Handle("GET /api/profile", middleware.AuthMiddleware(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims := r.Context().Value(middleware.UserContextKey).(*services.Claims)

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{
				"userId": claims.UserID,
				"email":  claims.Email,
			})
		}),
	))

	mux.Handle("POST /api/refresh", middleware.RateLmitMiddleWare(http.HandlerFunc(handlers.Refresh)))

	mux.Handle("POST /api/posts", middleware.AuthMiddleware(http.HandlerFunc(handlers.CreatePost)))

	mux.Handle("GET /api/posts", middleware.AuthMiddleware(
		http.HandlerFunc(handlers.GetPosts),
	))
	mux.Handle("GET /api/posts/{id}", middleware.AuthMiddleware(
		http.HandlerFunc(handlers.GetPost),
	))
	mux.Handle("DELETE /api/posts/{id}", middleware.AuthMiddleware(
		http.HandlerFunc(handlers.DeletePost),
	))

	log.Printf("Server starting on port 8080...")
	if err := http.ListenAndServe(":"+cfg.Port, middleware.LoggerMiddleWare(mux)); err != nil {
		slog.Error("Failed to start server",
			"error", err)
		os.Exit(1)
	}
}
