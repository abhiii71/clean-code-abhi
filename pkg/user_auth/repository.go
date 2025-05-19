package userauth

import (
	"context"
	"database/sql"
)

type Repository interface {
	Register(ctx context.Context, request UserRegisterRequest) error
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{
		db: db,
	}
}

func (r *repository) Register(ctx context.Context, request UserRegisterRequest) error {
	query := `INSERT INTO users (first_name, last_name, email, password) VALUES($1, $2, $3, $4)`

	_, err := r.db.Exec(query, request.FirstName, request.LastName, request.Email, request.Password)
	return err
}
