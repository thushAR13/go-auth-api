package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type brevoRequest struct {
	Sender      brevoContact   `json:"sender"`
	To          []brevoContact `json:"to"`
	Subject     string         `json:"subject"`
	TextContent string         `json:"textContent"`
}

type brevoContact struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

func SendWelcomeEmail(toEmail, toName string) error {
	payload := brevoRequest{
		Sender: brevoContact{
			Name:  senderName,
			Email: senderEmail,
		},
		To: []brevoContact{
			{Name: toName, Email: toEmail},
		},
		Subject:     "Welcome to go-auth-api",
		TextContent: fmt.Sprintf("Hey %s \n, Welcome to go-auth-api!\n\n Your account has been created succesfully!!", toName),
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("Failed to encode email payload: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, "https://api.brevo.com/v3/smtp/email", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to build email request: %w", err)
	}

	req.Header.Set("api-key", brevoAPIKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send email request %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		var brevError map[string]any
		json.NewDecoder(resp.Body).Decode(&brevError)

		return fmt.Errorf("brevo API returned unexpected status code: %d, %v", resp.StatusCode, brevError)
	}

	return nil
}
