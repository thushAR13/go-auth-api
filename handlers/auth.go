package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"go-auth-api/db"
	"go-auth-api/models"
	"go-auth-api/services"
	"log/slog"
	"net/http"
	"time"
)

func Register(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" || req.Email == "" || req.Password == "" {
		http.Error(w, "Name, Email and Password are required", http.StatusBadRequest)
		return
	}

	var exists bool
	err := db.DB.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM users WHERE email=$1)", req.Email).Scan(&exists)
	if err != nil {
		slog.Error("Error checking database",
			"error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if exists {
		http.Error(w, "Email already registered", http.StatusConflict)
		return
	}

	hashedPassword, err := services.HashPassword(req.Password)

	if err != nil {
		slog.Error("Failed to hash password",
			"error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	var user models.User

	err = db.DB.QueryRowContext(ctx, "INSERT INTO users (name, email, password) values ($1, $2, $3) RETURNING id, name, email",
		req.Name, req.Email, hashedPassword).Scan(&user.ID, &user.Name, &user.Email)

	if err != nil {
		slog.Error("Failed to print user",
			"error", err,
			"ID", user.ID,
			"email", user.Email)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if err = services.SendWelcomeEmail(req.Email, req.Name); err != nil {
		slog.Error("Error sending email",
			"error", err,
			"email", req.Email)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)

}

func Login(w http.ResponseWriter, r *http.Request) {

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Email == "" || req.Password == "" {
		http.Error(w, "Email and password are required", http.StatusBadRequest)
		return
	}

	var user models.User

	err := db.DB.QueryRow("SELECT id, name, email, password FROM users WHERE email = $1", req.Email).
		Scan(&user.ID, &user.Name, &user.Email, &user.Password)
	if err == sql.ErrNoRows {
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}
	if err != nil {
		slog.Error("Error frtching user",
			"error", err,
			"email", req.Email)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if err := services.CheckPassword(req.Password, user.Password); err != nil {
		http.Error(w, "Invalid email or password", http.StatusBadRequest)
		return
	}

	accessToken, err := services.GenerateToken(user.ID, user.Email)
	if err != nil {
		slog.Error("Token generation failed",
			"error", err,
			"email", user.Email)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	refreshToken, err := services.GenerateRefreshToken()
	if err != nil {
		slog.Error("Failed to generate refresh token",
			"error", err)
		http.Error(w, "Failed to generate refresh token", http.StatusInternalServerError)
		return
	}

	expiresAt := time.Now().Add(7 * 24 * time.Hour)
	if err := db.SaveRefreshToken(ctx, user.ID, refreshToken, expiresAt); err != nil {
		slog.Error("failed to save refresh token", "error", err, "userID", user.ID)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"token":         accessToken,
		"refresh_token": refreshToken,
		"expires_in":    15 * 60,
	})

}

func Refresh(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	slog.Info("Refresh Token", "refresh token", req.RefreshToken)

	if req.RefreshToken == "" {
		slog.Error("No refresh token found")
		http.Error(w, "No refresh token found", http.StatusBadRequest)
		return
	}
	var token models.RefreshToken
	err := db.DB.QueryRow("SELECT user_id, token, expires_at FROM refresh_tokens WHERE token = $1", req.RefreshToken).
		Scan(&token.UserId, &token.Token, &token.ExpiresAt)
	if err != nil {
		if err == sql.ErrNoRows {
			slog.Error("Refresh token not found in db", "error:", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		slog.Error("Refresh token db error", "error:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if token.ExpiresAt.Before(time.Now()) {
		slog.Error("Refresh token has expired")
		http.Error(w, "Login expired", http.StatusForbidden)
		return
	}

	var newRefreshToken string

	newRefreshToken, err = services.GenerateRefreshToken()
	db.SaveRefreshToken(ctx, token.UserId, newRefreshToken, time.Now().Add(7*24*time.Hour))
	db.DeleteRefreshToken(ctx, req.RefreshToken)

	var user models.User

	err = db.DB.QueryRow("SELECT email FROM users WHERE id=$1", token.UserId).Scan(&user.Email)
	if err != nil {
		slog.Error("Email not found for the user")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	accessToken, err := services.GenerateToken(token.UserId, user.Email)

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(map[string]any{
		"token":         accessToken,
		"refresh_token": newRefreshToken,
		"expires_in":    7 * 24 * 60 * 60,
	})

}
