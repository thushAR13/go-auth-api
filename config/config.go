package config

import (
	"fmt"
	"os"
)

type Config struct {
	Port        string
	DBHost      string
	DBUser      string
	DBPassword  string
	DBPort      string
	DBName      string
	JWTSecret   string
	BrevoAPIKey string
	SenderEmail string
	SenderName  string
}

func Load() (*Config, error) {
	cfg := &Config{
		Port:        getEnv("PORT", "8000"),
		DBHost:      getEnv("DB_HOST", "localhost"),
		DBUser:      getEnv("DB_USER", "postgres"),
		DBPort:      getEnv("DB_PORT", "5432"),
		DBPassword:  getEnv("DB_PASSWORD", ""),
		DBName:      getEnv("DB_NAME", "goauthapi"),
		JWTSecret:   getEnv("JWT_SECRET", ""),
		BrevoAPIKey: getEnv("BREVO_API_KEY", ""),
		SenderEmail: getEnv("SENDER_EMAIL", ""),
		SenderName:  getEnv("SENDER_NAME", "goauthapi"),
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) validate() error {
	required := map[string]string{
		"JWT_SECRET":    c.JWTSecret,
		"BREVO_API_KEY": c.BrevoAPIKey,
		"SENDER_EMAIL":  c.SenderEmail,
		"DB_PASSWORD":   c.DBPassword,
	}

	for key, value := range required {
		if value == "" {
			return fmt.Errorf("required environment variable not found: %s", key)
		}
	}
	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
