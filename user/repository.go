package user

import (
	"context"
	"database/sql"
)

type Repository struct {
	db *sql.DB
}

type User struct {
	ID         string `json:"id" validate:"required,uuid"`
	TokensUsed int    `json:"tokens_used"`
}

func (r *Repository) Get(ctx context.Context, userID string) (User, error) {
	row := r.db.QueryRowContext(ctx, "SELECT id, tokens_used FROM users WHERE id = $1", userID)
	return toUser(row)
}

func (r *Repository) AddTokensUsed(ctx context.Context, tokens int, userID string) (User, error) {
	row := r.db.QueryRowContext(ctx, "UPDATE users SET tokens_used = tokens_used + $1 WHERE id = $2 RETURNING id, tokens_used", tokens, userID)
	return toUser(row)
}

func toUser(row *sql.Row) (User, error) {
	var user User
	if err := row.Scan(&user.ID, &user.TokensUsed); err != nil {
		return User{}, err
	}
	return user, nil
}

func NewUserRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}
