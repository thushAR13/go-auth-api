package db

import (
	"context"
	"time"
)

func (s *Store) SaveRefreshToken(ctx context.Context, userId int, token string, expiresAt time.Time) error {
	_, err := s.DB.ExecContext(ctx, "INSERT INTO refresh_tokens (user_id, token, expires_at) VALUES ($1, $2, $3)", userId, token, expiresAt)
	return err

}

func (s *Store) GetRefreshToken(ctx context.Context, token string) (int, error) {
	var userId int
	err := s.DB.QueryRowContext(ctx, "SELECT user_if FROM refresh_tokens WHERE token = $1 and expires_at > NOW()", token).Scan(&userId)
	return userId, err
}

func (s *Store) DeleteRefreshToken(ctx context.Context, token string) error {
	_, err := s.DB.ExecContext(ctx, "DELETE FROM refresh_tokens WHERE token=$1", token)
	return err
}

func (s *Store) DeleteAllUserRefreshTokens(ctx context.Context, userId int) error {
	_, err := s.DB.ExecContext(ctx, "DELETE FROM refresh_tokens WHERE user_id=$1", userId)
	return err
}
