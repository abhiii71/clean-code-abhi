package userauth

import (
	"context"
	"database/sql"
)

type Repository interface {
	Register(ctx context.Context, request UserRegisterRequest) (string, error)
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{
		db: db,
	}
}

func (r *repository) Register(ctx context.Context, request UserRegisterRequest) (string, error) {
	query := `INSERT INTO users (first_name, last_name, email, password) VALUES($1, $2, $3, $4) RETURNING id`

	var userID string
	err := r.db.QueryRowContext(ctx, query, request.FirstName, request.LastName, request.Email, request.Password).Scan(&userID)
	return userID, err
}
