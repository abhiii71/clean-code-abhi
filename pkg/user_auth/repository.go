package userauth

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"time"
)

type Repository interface {
	UserRegister(ctx context.Context, request UserRegisterRequest) (string, error)
	GetUserProfile(ctx context.Context, email string) (User, error)
	FindByEmail(ctx context.Context, email string) (User, error)
	FindByUserID(ctx context.Context, userID string) (User, error)
	InsertUserInformation(ctx context.Context, request UserInformationRequest) error
	UpdateUserInformation(ctx context.Context, userInfo UserInformationRequest) error

	FindByUUID(ctx context.Context, request UserInformationRequest) (UserInformationRequest, error)
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

// update user information

func (r *repository) UpdateUserInformation(ctx context.Context, req UserInformationRequest) error {
	query := `
        INSERT INTO user_information (user_uuid, id, address, vehicle, created_at, updated_at)
        VALUES ($1, $2, $3, $4, NOW(), NOW())
        ON CONFLICT (user_uuid) DO UPDATE
        SET address = EXCLUDED.address,
            vehicle = EXCLUDED.vehicle,
            updated_at = NOW()
    `

	addressJSON, err := json.Marshal(req.Address)
	if err != nil {
		return err
	}

	vehicleJSON, err := json.Marshal(req.Vehicle)
	if err != nil {
		return err
	}

	_, err = r.db.ExecContext(ctx, query, req.UUID, req.ID, addressJSON, vehicleJSON)
	return err
}

// user_information
func (r *repository) InsertUserInformation(ctx context.Context, req UserInformationRequest) error {
	addressJSON, err := json.Marshal(req.Address)
	if err != nil {
		log.Printf("[InsertUserInformation] Error marshaling address: %v\n", err)
		return err
	}
	log.Printf("[InsertUserInformation] Marshaled Address: %s\n", string(addressJSON))

	vehicleJSON, err := json.Marshal(req.Vehicle)
	if err != nil {
		log.Printf("[InsertUserInformation] Error marshaling vehicle: %v\n", err)
		return err
	}
	log.Printf("[InsertUserInformation] Marshaled Vehicle: %s\n", string(vehicleJSON))

	query := `INSERT INTO user_info (id, address, vehicle) VALUES ($1, $2, $3)`
	_, err = r.db.ExecContext(ctx, query, req.ID, addressJSON, vehicleJSON)
	log.Printf("[InsertUserInformation] DB Exec Error: %v\n", err)
	return err
}

func (r *repository) FindByUUID(ctx context.Context, request UserInformationRequest) (UserInformationRequest, error) {
	var user UserInformationRequest

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	query := `SELECT uuid, id, address, vehicle FROM user_information WHERE uuid = $1`
	err := r.db.QueryRowContext(ctx, query, request.UUID).Scan(
		&user.UUID,
		&user.ID,
		&user.Address,
		&user.Vehicle,
	)
	if err != nil {
		return UserInformationRequest{}, err
	}
	return user, nil
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
