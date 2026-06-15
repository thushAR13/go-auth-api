package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"go-auth-api/db"
	"go-auth-api/models"
	"go-auth-api/services"
	"log/slog"
	"net/http"
	"time"
)

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
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

	hashedPassword, err := services.HashPassword(req.Password)
	if err != nil {
		slog.Error("failed to hash password", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	user, err := h.store.CreateUser(ctx, req.Name, req.Email, hashedPassword)
	if err != nil {
		if errors.Is(err, db.ErrAlreadyExists) {
			http.Error(w, "email already registered", http.StatusConflict)
			return
		}
		slog.Error("Register: unexpected error", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
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

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {

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

	user, err := h.store.GetUserByEmail(ctx, req.Email)
	if errors.Is(err, db.ErrNotFound) {
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}
	if err != nil {
		slog.Error("Error fetching user",
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
	if err := h.store.SaveRefreshToken(ctx, user.ID, refreshToken, expiresAt); err != nil {
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

func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
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

	token, err := h.store.GetRefreshToken(ctx, req.RefreshToken)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			slog.Error("Refresh token not found in db", "error:", db.ErrNotFound)
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
	err = h.store.RotateRefreshToken(ctx, req.RefreshToken, token.UserId, newRefreshToken, time.Now().Add(7*24*time.Hour))

	if err != nil {
		slog.Error("Error rotating refresh token", "Error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	user, err := h.store.GetUserById(ctx, token.UserId)
	if err != nil {
		slog.Error("User fetch error - ID", "Error", err)
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

func (h *Handler) Profile(w http.ResponseWriter, r *http.Request) {
	claims := getClaimsFromContext(r)
	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(map[string]any{
		"userId": claims.UserID,
		"email":  claims.Email,
	})
}
