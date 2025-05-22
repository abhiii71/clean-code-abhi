package userauth

import (
	"context"
	"database/sql"
	"log"
	"time"
)

type Repository interface {
	UserRegister(ctx context.Context, request UserRegisterRequest) (string, error)
	GetUserProfile(ctx context.Context, email string) (User, error)
	FindByEmail(ctx context.Context, email string) (User, error)
	FindByUserID(ctx context.Context, userID string) (User, error)
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{
		db: db,
	}
}

func (r *repository) UserRegister(ctx context.Context, request UserRegisterRequest) (string, error) {
	query := `INSERT INTO users 
    (first_name, last_name, email, password, dob, gender)
    VALUES ($1, $2, $3, $4, $5, $6)
    RETURNING id`
	var userID string
	err := r.db.QueryRowContext(
		ctx,
		query,
		request.FirstName,
		request.LastName,
		request.Email,
		request.Password,
		time.Time(request.DOB), // Convert custom Date to time.Time
		request.Gender,
	).Scan(&userID)

	if err != nil {
		log.Println("error from repo", err.Error())
		return "", err
	}
	return userID, nil
}

func (r *repository) FindByEmail(ctx context.Context, email string) (User, error) {
	var user User

	query := "SELECT id, email, password FROM users WHERE email=$1"
	row := r.db.QueryRowContext(ctx, query, email)

	err := row.Scan(&user.ID, &user.Email, &user.Password)

	if err != nil {
		return User{}, err
	}
	return user, nil
}

func (r *repository) FindByUserID(ctx context.Context, userId string) (User, error) {
	var user User

	query := "SELECT id, first_name, last_name, email, password, dob, gender FROM users WHERE id=$1"

	row := r.db.QueryRowContext(ctx, query, userId)

	err := row.Scan(&user.ID, &user.FirstName, &user.LastName, &user.Email, &user.Password, &user.DOB, &user.Gender)
	if err != nil {
		log.Println("[FindByUserID] DB scan error:", err)
		return User{}, err
	}
	return user, nil
}

func (r *repository) GetUserProfile(ctx context.Context, email string) (User, error) {
	var user User
	query := "SELECT id, email, password FROM users WHERE email=$1"
	row := r.db.QueryRowContext(ctx, query, email)

	err := row.Scan(&user.ID, &user.Email, &user.Password)
	// log.Println("[DB password retrieved]:", user.Password)
	if err != nil {
		return User{}, err
	}
	return user, nil

}
