package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"go-auth-api/models"
	"time"
)

func (s *Store) SaveRefreshToken(ctx context.Context, userId int, token string, expiresAt time.Time) error {
	_, err := s.DB.ExecContext(ctx, "INSERT INTO refresh_tokens (user_id, token, expires_at) VALUES ($1, $2, $3)", userId, token, expiresAt)
	return err

}

func (s *Store) GetRefreshToken(ctx context.Context, token string) (*models.RefreshToken, error) {
	var refreshToken models.RefreshToken
	err := s.DB.QueryRowContext(ctx, "SELECT user_id, token, expires_at FROM refresh_tokens WHERE token = $1 and expires_at > NOW()", token).
		Scan(&refreshToken.UserId, &refreshToken.Token, &refreshToken.ExpiresAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &refreshToken, nil
}

func (s *Store) DeleteRefreshToken(ctx context.Context, token string) error {
	_, err := s.DB.ExecContext(ctx, "DELETE FROM refresh_tokens WHERE token=$1", token)
	return err
}

func (s *Store) DeleteAllUserRefreshTokens(ctx context.Context, userId int) error {
	_, err := s.DB.ExecContext(ctx, "DELETE FROM refresh_tokens WHERE user_id=$1", userId)
	return err
}

func (s *Store) RotateRefreshToken(ctx context.Context, oldToken string, userID int, newToken string, expiresAt time.Time) error {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("RotateRefereshToken: begin: %w", err)
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, "INSERT INTO refresh_tokens (user_id, token, expires_at) VALUES ($1, $2, $3)",
		userID, newToken, expiresAt)
	if err != nil {
		return fmt.Errorf("RotateRefreshToke insert: %w", err)
	}
	_, err = tx.ExecContext(ctx, "DELETE FROM refresh_tokens WHERE token=$1", oldToken)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("RotateRefreshToken delete: %w", err)
	}

	return tx.Commit()

}
