package userauth

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"strings"
	"time"
)

type Repository interface {
	UserRegister(ctx context.Context, request UserRegisterRequest) (string, error)
	GetUserProfile(ctx context.Context, email string) (*User, error)
	FindUserByEmail(ctx context.Context, email string) (*User, error)
	FindUserByID(ctx context.Context, userID string) (*User, []byte, []byte, error)
	// (User, error)
	// InsertUserInformation(ctx context.Context, request UserInformationRequest) error
	UpdateUserInfo(ctx context.Context, request UserInformationRequest) error

	// FindByUUID(ctx context.Context, request UserInformationRequest) (UserInformationRequest, error)
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
func (r *repository) UpdateUserInfo(ctx context.Context, request UserInformationRequest) error {
	addressJSON, err := json.Marshal(request.Address)
	if err != nil {
		log.Println("Failed to marshal address:", err)
		return err
	}

	vehicleJSON, err := json.Marshal(request.Vehicle)
	if err != nil {
		log.Println("Failed to marshal vehicle:", err)
		return err
	}

	query := `
        INSERT INTO user_information (id, address, vehicle, created_at, updated_at)
        VALUES ($1, $2, $3, NOW(), NOW())
        ON CONFLICT (id) DO UPDATE
        SET 
            address = COALESCE(EXCLUDED.address, user_information.address),
            vehicle = COALESCE(EXCLUDED.vehicle, user_information.vehicle),
            updated_at = NOW()
    `

	_, err = r.db.ExecContext(ctx, query, request.ID, addressJSON, vehicleJSON)
	if err != nil {
		log.Println("[UpsertUserInformation] Error executing query:", err)
		return err
	}
	return nil
}

// Find user by Email
func (r *repository) FindUserByEmail(ctx context.Context, email string) (*User, error) {
	var user User
	query := "SELECT id, email, password FROM users WHERE LOWER(email) = $1"
	row := r.db.QueryRowContext(ctx, query, strings.ToLower(email))
	err := row.Scan(&user.ID, &user.Email, &user.Password)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Printf("[FindUserByEmail] No user found for email: %s\n", email)
			return nil, errors.New("user not found")
		}
		log.Printf("[FindUserByEmail] DB error: %v\n", err)
		return nil, errors.New("internal database error")
	}
	return &user, nil
}

func (r *repository) FindUserByID(ctx context.Context, userID string) (*User, []byte, []byte, error) {
	var user User
	var addressBytes, vehicleBytes []byte

	query := `
    SELECT 
      u.id, u.email, u.password, u.dob, u.first_name, u.last_name, u.gender,
      ui.address, ui.vehicle
    FROM users u
    LEFT JOIN user_information ui ON u.id = ui.id
    WHERE u.id = $1;
    `

	row := r.db.QueryRowContext(ctx, query, userID)

	err := row.Scan(
		&user.ID,
		&user.Email,
		&user.Password,
		&user.DOB,
		&user.FirstName,
		&user.LastName,
		&user.Gender,
		&addressBytes,
		&vehicleBytes,
	)
	if err != nil {
		log.Println("[GetProfile] Error scanning user:", err)
		return nil, nil, nil, err
	}

	return &user, addressBytes, vehicleBytes, nil
}

func (r *repository) GetUserProfile(ctx context.Context, email string) (*User, error) {
	var user User
	var addressBytes, vehicleBytes []byte
	query := `
SELECT 
    u.id, u.email, u.password,
    ui.address, ui.vehicle, ui.dob, ui.first_name, ui.last_name, ui.gender
FROM 
    users u
LEFT JOIN 
    user_information ui ON u.id = ui.id
WHERE 
    u.id = $1
`
	row := r.db.QueryRowContext(ctx, query, user.ID)

	// err := row.Scan(&user.ID, &user.Email, &user.Password)
	err := row.Scan(
		&user.ID, &user.Email, &user.Password,
		&user.DOB, &user.FirstName, &user.LastName, &user.Gender, &addressBytes, &vehicleBytes,

		// log.Printf("[GetProfile] Scan error: %v", err)
	)
	// log.Println("[DB password retrieved]:", user.Password)
	if err != nil {
		log.Printf("[GetProfile] Executing query for user ID: %s", user.ID)
		log.Printf("[GetProfile] Query: %s", query)
		return nil, err
	}
	return &user, nil

}
