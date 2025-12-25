package database

import (
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
	Password  string    `json:"-"` // Never expose password in JSON
	FullName  string    `json:"full_name"`
}

type CreateUserParams struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	FullName string `json:"full_name"`
}

func (c Client) GetUsers() ([]User, error) {
	query := `
		SELECT
			id,
			email,
			full_name
		FROM users
	`

	rows, err := c.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := []User{}
	for rows.Next() {
		var user User
		var id string
		if err := rows.Scan(&id, &user.Email, &user.FullName); err != nil {
			return nil, err
		}
		user.ID, err = uuid.Parse(id)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}

func (c Client) GetUserByEmail(email string) (User, error) {
	query := `
		SELECT id, created_at, updated_at, email, password, full_name
		FROM users
		WHERE email = ?
	`
	var user User
	var id string
	err := c.db.QueryRow(query, email).Scan(&id, &user.CreatedAt, &user.UpdatedAt, &user.Email, &user.Password, &user.FullName)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return User{}, nil
		}
		return User{}, err
	}
	user.ID, err = uuid.Parse(id)
	if err != nil {
		return User{}, err
	}
	return user, nil
}

func (c Client) GetUserByRefreshToken(token string) (*User, error) {
	query := `
		SELECT u.id, u.email, u.created_at, u.updated_at, u.password, u.full_name
		FROM users u
		JOIN refresh_tokens rt ON u.id = rt.user_id
		WHERE rt.token = ?
		  AND rt.revoked_at IS NULL
		  AND rt.expires_at > CURRENT_TIMESTAMP
	`

	var user User
	var id string
	err := c.db.QueryRow(query, token).Scan(&id, &user.Email, &user.CreatedAt, &user.UpdatedAt, &user.Password, &user.FullName)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	user.ID, err = uuid.Parse(id)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (c Client) CreateUser(params CreateUserParams) (*User, error) {
	id := uuid.New()

	query := `
		INSERT INTO users
		    (id, created_at, updated_at, email, password, full_name)
		VALUES
		    (?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, ?, ?, ?)
	`
	_, err := c.db.Exec(query, id.String(), params.Email, params.Password, params.FullName)
	if err != nil {
		return nil, err
	}

	return c.GetUser(id)
}

func (c Client) GetUser(id uuid.UUID) (*User, error) {
	query := `
		SELECT id, created_at, updated_at, email, password, full_name
		FROM users
		WHERE id = ?
	`
	var user User
	var idStr string
	err := c.db.QueryRow(query, id.String()).Scan(&idStr, &user.CreatedAt, &user.UpdatedAt, &user.Email, &user.Password, &user.FullName)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	user.ID, err = uuid.Parse(idStr)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (c Client) UpdateUserPassword(userID uuid.UUID, hashedPassword string) error {
	query := `
		UPDATE users
		SET password = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`
	_, err := c.db.Exec(query, hashedPassword, userID.String())
	return err
}

func (c Client) DeleteUser(id uuid.UUID) error {
	query := `
		DELETE FROM users
		WHERE id = ?
	`
	_, err := c.db.Exec(query, id.String())
	return err
}
