package handlers

import (
	"context"
	"go-auth-api/models"
	"time"
)

type Storer interface {
	GetPostsByUserId(ctx context.Context, userId int) ([]models.Post, error)
	GetPostByID(ctx context.Context, postID int) (*models.Post, error)
	DeletePost(ctx context.Context, postID, userID int) error
	CreatePost(ctx context.Context, userID int, title, body string) (*models.Post, error)
	CreateUser(ctx context.Context, name string, email string, password string) (*models.User, error)
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	SaveRefreshToken(ctx context.Context, userId int, token string, expiresAt time.Time) error
	GetRefreshToken(ctx context.Context, token string) (*models.RefreshToken, error)
	DeleteRefreshToken(ctx context.Context, token string) error
	GetUserById(ctx context.Context, id int) (*models.User, error)
}
