package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"go-auth-api/models"
)

func (s *Store) CreateUser(ctx context.Context, name, email, password string) (*models.User, error) {
	var exists bool
	err := s.DB.QueryRowContext(ctx,
		"SELECT EXISTS(SELECT 1 FROM users WHERE email=$1)", email,
	).Scan(&exists)
	if err != nil {
		return nil, fmt.Errorf("CreateUser: check exists: %w", err)
	}

	if exists {
		return nil, ErrAlreadyExists
	}

	var user models.User
	err = s.DB.QueryRowContext(ctx,
		"INSERT INTO users (name, email, password) VALUES($1, $2, $3) RETURNING id, name, email",
		name, email, password,
	).Scan(&user.ID, &user.Name, &user.Email)
	if err != nil {
		return nil, fmt.Errorf("CreateUser: insert: %w", err)
	}

	return &user, nil
}

func (s *Store) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := s.DB.QueryRowContext(ctx, "SELECT id, name, email, password FROM users WHERE email=$1", email).
		Scan(&user.ID, &user.Name, &user.Email, &user.Password)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("Failed fetching user: %w", err)
	}
	return &user, nil
}

func (s *Store) GetUserById(ctx context.Context, id int) (*models.User, error) {
	var user models.User
	err := s.DB.QueryRowContext(ctx, "SELECT id, name, email, password FROM users WHERE id=$1", id).
		Scan(&user.ID, &user.Name, &user.Email, &user.Password)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("Failed fetching user: %w", err)
	}
	return &user, nil
}
