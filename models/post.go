package models

import "time"

type Post struct {
	ID        int       `json:"id"`
	UserID    int       `json:"userId"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"createdAt"`
}

type CreatePostRequest struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}
