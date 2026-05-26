package db

import (
	"context"
	"time"
)

func SaveRefreshToken(ctx context.Context, userId int, token string, expiresAt time.Time) error {
	_, err := DB.ExecContext(ctx, "INSERT INTO refresh_tokens (user_id, token, expires_at) VALUES ($1, $2, $3)", userId, token, expiresAt)
	return err

}

func GetRefreshToken(ctx context.Context, token string) (int, error) {
	var userId int
	err := DB.QueryRowContext(ctx, "SELECT user_if FROM refresh_tokens WHERE token = $1 and expires_at > NOW()", token).Scan(&userId)
	return userId, err
}

func DeleteRefreshToken(ctx context.Context, token string) error {
	_, err := DB.ExecContext(ctx, "DELETE FROM refresh_tokens WHERE token=$1", token)
	return err
}

func DeleteAllUserRefreshTokens(ctx context.Context, userId int) error {
	_, err := DB.ExecContext(ctx, "DELETE FROM refresh_tokens WHERE user_id=$1", userId)
	return err
}
