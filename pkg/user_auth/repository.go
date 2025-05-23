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
	FindByUserID(ctx context.Context, userID string) (User, *json.RawMessage, *json.RawMessage, error)
	// (User, error)
	// InsertUserInformation(ctx context.Context, request UserInformationRequest) error
	UpsertUserInformation(ctx context.Context, userInfo UserInformationRequest) error

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
func (r *repository) UpsertUserInformation(ctx context.Context, req UserInformationRequest) error {
	// if req.UUID == uuid.Nil {
	// 	req.UUID = uuid.New()
	// }

	addressJSON, err := json.Marshal(req.Address)
	if err != nil {
		return err
	}

	vehicleJSON, err := json.Marshal(req.Vehicle)
	if err != nil {
		return err
	}

	query := `
        INSERT INTO user_information (id, address, vehicle, created_at, updated_at)
        VALUES ($1, $2, $3, NOW(), NOW())
        ON CONFLICT (id) DO UPDATE
        SET 
            address = EXCLUDED.address,
            vehicle = EXCLUDED.vehicle,
            updated_at = NOW()
    `

	_, err = r.db.ExecContext(ctx, query, req.ID, addressJSON, vehicleJSON)
	return err
}

// func (r *repository) UpdateUserInformation(ctx context.Context, req UserInformationRequest) error {
// 	// Marshal Vehicle struct to JSON bytes
// 	vehicleJSON, err := json.Marshal(req.Vehicle)
// 	if err != nil {
// 		return err
// 	}

// 	// Marshal Address struct to JSON bytes
// 	// Since Address fields are json.RawMessage, marshal will convert the entire struct to JSON object
// 	addressJSON, err := json.Marshal(req.Address)
// 	if err != nil {
// 		return err
// 	}

// 	query := `
// 		UPDATE user_information
// 		SET address = $1,
// 		    vehicle = $2,
// 		    updated_at = NOW()
// 		WHERE id = $3
// 	`

// 	_, err = r.db.ExecContext(ctx, query, addressJSON, vehicleJSON, req.ID)
// 	return err
// }

// // user_information
// func (r *repository) InsertUserInformation(ctx context.Context, req UserInformationRequest) error {
// 	addressJSON, err := json.Marshal(req.Address)
// 	if err != nil {
// 		log.Printf("[InsertUserInformation] Error marshaling address: %v\n", err)
// 		return err
// 	}
// 	log.Printf("[InsertUserInformation] Marshaled Address: %s\n", string(addressJSON))

// 	vehicleJSON, err := json.Marshal(req.Vehicle)
// 	if err != nil {
// 		log.Printf("[InsertUserInformation] Error marshaling vehicle: %v\n", err)
// 		return err
// 	}
// 	log.Printf("[InsertUserInformation] Marshaled Vehicle: %s\n", string(vehicleJSON))

// 	query := `INSERT INTO user_info (id, address, vehicle) VALUES ($1, $2, $3)`
// 	_, err = r.db.ExecContext(ctx, query, req.ID, addressJSON, vehicleJSON)
// 	log.Printf("[InsertUserInformation] DB Exec Error: %v\n", err)
// 	return err
// }

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

func (r *repository) FindByUserID(ctx context.Context, userID string) (User, *json.RawMessage, *json.RawMessage, error) {
	var user User

	var addressJSON *json.RawMessage
	var vehicleJSON *json.RawMessage

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
		&addressJSON, //uint8
		&vehicleJSON, //uint8
	)
	if err != nil {
		log.Println("[GetProfile] Error scanning user:", err)
		return user, nil, nil, err
	}

	return user, addressJSON, vehicleJSON, nil
}

func (r *repository) GetUserProfile(ctx context.Context, email string) (User, error) {
	var user User
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
		&user.Address, &user.Vehicle, &user.DOB, &user.FirstName, &user.LastName, &user.Gender,

		// log.Printf("[GetProfile] Scan error: %v", err)
	)
	// log.Println("[DB password retrieved]:", user.Password)
	if err != nil {
		log.Printf("[GetProfile] Executing query for user ID: %s", user.ID)
		log.Printf("[GetProfile] Query: %s", query)
		return User{}, err
	}
	return user, nil

}
