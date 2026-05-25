package db

import (
	"time"
)

func SaveRefreshToken(userId int, token string, expiresAt time.Time) error {
	_, err := DB.Exec("INSERT INTO refresh_tokens (user_id, token, expires_at) VALUES ($1, $2, $3)", userId, token, expiresAt)
	return err

}

func GetRefreshToken(token string) (int, error) {
	var userId int
	err := DB.QueryRow("SELECT user_if FROM refresh_tokens WHERE token = $1 and expires_at > NOW()", token).Scan(&userId)
	return userId, err
}

func DeleteRefreshToken(token string) error {
	_, err := DB.Exec("DELETE FROM refresh_tokens WHERE token=$1", token)
	return err
}

func DeleteAllUserRefreshTokens(userId int) error {
	_, err := DB.Exec("DELETE FROM refresh_tokens WHERE user_id=$1", userId)
	return err
}
