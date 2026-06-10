package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"go-auth-api/middleware"
	"go-auth-api/models"
	"go-auth-api/services"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func getClaimsFromContext(r *http.Request) *services.Claims {
	return r.Context().Value(middleware.UserContextKey).(*services.Claims)
}

func (h *Handler) CreatePost(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	claims := getClaimsFromContext(r)

	var req models.CreatePostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("Invalid request body", "error", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Title == "" || req.Body == "" {
		slog.Error("Body and title are empty")
		http.Error(w, "Title and body are required", http.StatusBadRequest)
		return
	}

	post, err := h.store.CreatePost(ctx, claims.UserID, req.Title, req.Body)
	if err != nil {
		slog.Error("failed to create post", "error", err, "userID", claims.UserID)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(post)

}

func (h *Handler) GetPosts(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	claims := getClaimsFromContext(r)
	posts, err := h.store.GetPostsByUserId(ctx, claims.UserID)
	if err != nil {
		slog.Error("failed to fatch posts", "error", err, "UserID", claims.UserID)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(posts)
}

func (h *Handler) GetPost(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	claims := getClaimsFromContext(r)
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	post, err := h.store.GetPostByID(ctx, id)
	if err == sql.ErrNoRows {
		slog.Error("Post not found", "Error", err)
		http.Error(w, "Post not found", http.StatusNotFound)
		return
	}
	if err != nil {
		slog.Error("Failed to fetch post", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if post.UserID != claims.UserID {
		http.Error(w, "Post not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(post)
}

func (h *Handler) DeletePost(w http.ResponseWriter, r *http.Request) {
	claims := getClaimsFromContext(r)
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Invalid post id", http.StatusBadRequest)
		return
	}

	if err := h.store.DeletePost(ctx, id, claims.UserID); err != nil {
		if strings.Contains(err.Error(), "Post not found") {
			http.Error(w, "Post not found", http.StatusNotFound)
			return
		}
		slog.Error("failed to delete the post", "error", err, "postID", id, "UserID", claims.UserID)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return

	}
	w.WriteHeader(http.StatusNoContent)
}
