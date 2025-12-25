package database

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"time"

	"github.com/google/uuid"
)

type PasswordResetToken struct {
	Token     string     `json:"token"`
	UserID    uuid.UUID  `json:"user_id"`
	ExpiresAt time.Time  `json:"expires_at"`
	UsedAt    *time.Time `json:"used_at"`
	CreatedAt time.Time  `json:"created_at"`
}

// CreatePasswordResetToken creates a new password reset token valid for 1 hour
func (c Client) CreatePasswordResetToken(userID uuid.UUID) (*PasswordResetToken, error) {
	// Generate random token
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return nil, err
	}
	token := hex.EncodeToString(tokenBytes)

	expiresAt := time.Now().UTC().Add(time.Hour) // Valid for 1 hour

	query := `
		INSERT INTO password_reset_tokens (token, user_id, expires_at, created_at)
		VALUES (?, ?, ?, CURRENT_TIMESTAMP)
	`
	_, err := c.db.Exec(query, token, userID.String(), expiresAt)
	if err != nil {
		return nil, err
	}

	return &PasswordResetToken{
		Token:     token,
		UserID:    userID,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now().UTC(),
	}, nil
}

// GetPasswordResetToken retrieves a valid (not expired, not used) password reset token
func (c Client) GetPasswordResetToken(token string) (*PasswordResetToken, error) {
	query := `
		SELECT token, user_id, expires_at, used_at, created_at
		FROM password_reset_tokens
		WHERE token = ?
		  AND used_at IS NULL
		  AND expires_at > CURRENT_TIMESTAMP
	`

	var prt PasswordResetToken
	var userIDStr string
	err := c.db.QueryRow(query, token).Scan(
		&prt.Token,
		&userIDStr,
		&prt.ExpiresAt,
		&prt.UsedAt,
		&prt.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // Token not found or expired/used
		}
		return nil, err
	}

	prt.UserID, err = uuid.Parse(userIDStr)
	if err != nil {
		return nil, err
	}

	return &prt, nil
}

// MarkPasswordResetTokenUsed marks a password reset token as used
func (c Client) MarkPasswordResetTokenUsed(token string) error {
	query := `
		UPDATE password_reset_tokens
		SET used_at = CURRENT_TIMESTAMP
		WHERE token = ?
	`
	_, err := c.db.Exec(query, token)
	return err
}

// DeleteExpiredPasswordResetTokens cleans up expired tokens
func (c Client) DeleteExpiredPasswordResetTokens() error {
	query := `
		DELETE FROM password_reset_tokens
		WHERE expires_at < CURRENT_TIMESTAMP OR used_at IS NOT NULL
	`
	_, err := c.db.Exec(query)
	return err
}
