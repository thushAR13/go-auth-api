package main

import (
	"context"
	"go-auth-api/config"
	"go-auth-api/db"
	"go-auth-api/handlers"
	"go-auth-api/middleware"
	"go-auth-api/services"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		slog.Warn("no .env file found, reading from environment")
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)
	slog.Info("Logger initialized")

	cfg, err := config.Load()
	if err != nil {
		slog.Error("configuration error", "error", err)
		os.Exit(1)
	}

	database := db.Init(cfg)
	store := db.NewStore(database)
	h := handlers.NewHandler(store)
	services.Init(cfg)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Welcome to go-auth-api!\n"))
	})

	// Public routes — rate limited only
	mux.Handle("POST /api/register", middleware.Chain(
		h.Register,
		middleware.RateLimitMiddleware,
	))
	mux.Handle("POST /api/login", middleware.Chain(
		h.Login,
		middleware.RateLimitMiddleware,
	))

	// Refresh and logout — no rate limit needed
	mux.HandleFunc("POST /api/refresh", h.Refresh)

	// Profile — auth only
	mux.Handle("GET /api/profile", middleware.AuthMiddleware(http.HandlerFunc(h.Profile)))

	// Post routes — auth + rate limited
	mux.Handle("POST /api/posts", middleware.Chain(
		h.CreatePost,
		middleware.AuthMiddleware,
		middleware.RateLimitMiddleware,
	))
	mux.Handle("GET /api/posts", middleware.Chain(
		h.GetPosts,
		middleware.AuthMiddleware,
		middleware.RateLimitMiddleware,
	))
	mux.Handle("GET /api/posts/{id}", middleware.Chain(
		h.GetPost,
		middleware.AuthMiddleware,
		middleware.RateLimitMiddleware,
	))
	mux.Handle("DELETE /api/posts/{id}", middleware.Chain(
		h.DeletePost,
		middleware.AuthMiddleware,
		middleware.RateLimitMiddleware,
	))

	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      middleware.LoggerMiddleware(mux),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine so main can listen for signals below
	go func() {
		slog.Info("server starting", "port", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server failed to start", "error", err)
			os.Exit(1)
		}
	}()

	// Block until Ctrl+C or kill signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutdown signal received, draining requests...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		slog.Error("server forced to shutdown", "error", err)
		os.Exit(1)
	}

	slog.Info("server shutdown cleanly")
}
