package db

import (
	"database/sql"
	"fmt"
	"go-auth-api/models"
)

func CreatePost(userID int, title, body string) (*models.Post, error) {
	post := &models.Post{}
	err := DB.QueryRow(
		`INSERT INTO posts (user_id, title, body)
		VALUES ($1, $2, $3)
		RETURNING id, user_id, title, body, created_at`,
		userID, title, body,
	).Scan(&post.ID, &post.UserId, &post.Title, &post.Body, &post.CreatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create post %w", err)
	}
	return post, nil
}

func GetPostsByUserId(userId int) ([]models.Post, error) {
	rows, err := DB.Query(
		`SELECT id, user_id, title, body, created_at
		FROM posts WHERE user_id = $1  ORDER BY created_at DESC`, userId,
	)
	if err != nil {
		return nil, fmt.Errorf("Failed to fetch posts: %w", err)
	}
	defer rows.Close()

	posts := []models.Post{}
	var post models.Post
	for rows.Next() {
		if err := rows.Scan(&post.ID, &post.UserId, &post.Title, &post.Body, &post.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan post: %w", err)
		}
		posts = append(posts, post)
	}
	return posts, nil
}

func GetPostByID(postID int) (*models.Post, error) {
	post := &models.Post{}

	err := DB.QueryRow(`SELECT id, user_id, title, body, created_at FROM posts WHERE id=$1`, postID).Scan(
		&post.ID, &post.UserId, &post.Title, &post.Body, &post.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, sql.ErrNoRows
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get post %w", err)
	}
	return post, nil
}

func DeletePost(postID, userID int) error {
	result, err := DB.Exec(
		`DELETE FROM posts WHERE id=$1 AND user_id=$2`, postID, userID,
	)
	if err != nil {
		return fmt.Errorf("failed to delete post %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("Post not found")
	}
	return nil
}
