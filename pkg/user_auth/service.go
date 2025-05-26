package userauth

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

type Service interface {
	UserRegister(ctx context.Context, request UserRegisterRequest) (string, error)
	GetUserProfile(ctx context.Context, request UserLoginRequest) (string, error)
	GetProfile(ctx context.Context, userId string) (User, error)
	// GetUserInfo(ctx context.Context, request UserInformationRequest) error
	UpdateUserInfo(ctx context.Context, req UserInformationRequest) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{
		repo: repo,
	}
}

func (s *service) UserRegister(ctx context.Context, request UserRegisterRequest) (string, error) {
	// business login
	// password hash
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Println("[hashed password Error]:", err)
		return "", err
	}

	_, err = s.repo.FindUserByEmail(ctx, request.Email)
	if err == nil {
		// log.Println("[FindByEmail Error]:", err)
		return "", fmt.Errorf("email already regestered")
	} else if !errors.Is(err, sql.ErrNoRows) {
		return "", fmt.Errorf("error checking existing email: %w", err)
	}
	request.Password = string(hashedPassword)

	// store the user
	userID, err := s.repo.UserRegister(ctx, request)
	if err != nil {
		return "", err
	}

	// user exists
	return userID, nil
}

var ErrNoRowsAffected = errors.New("no rows affected")

func (s *service) UpdateUserInfo(ctx context.Context, request UserInformationRequest) error {
	// Try to update first
	err := s.repo.UpdateUserInfo(ctx, request)
	return err
}

// User information update

// Helper functions to check empty structs
func isEmptyAddress(addr Address) bool {
	return addr.City == nil && addr.State == nil && addr.PostalCode == nil && addr.Country == nil
}

func isEmptyVehicle(v Vehicle) bool {
	// Customize depending on what empty means
	return (v.Car == nil && v.Bike == nil)
}

func (s *service) UpdateUserInformation(ctx context.Context, req UserInformationRequest) error {
	// Fetch existing data
	existingUser, addressBytes, vehicleBytes, err := s.repo.FindUserByID(ctx, fmt.Sprintf("%d", req.ID))
	if err != nil {
		return fmt.Errorf("user not found or error fetching: %w", err)
	}

	// Unmarshal existing address and vehicle
	if addressBytes != nil {
		if err := json.Unmarshal(addressBytes, &existingUser.Address); err != nil {
			return fmt.Errorf("failed to unmarshal existing address: %w", err)
		}
	}
	if vehicleBytes != nil {
		if err := json.Unmarshal(vehicleBytes, &existingUser.Vehicle); err != nil {
			return fmt.Errorf("failed to unmarshal existing vehicle: %w", err)
		}
	}

	// Merge Address fields
	if req.Address.City == nil {
		req.Address.City = existingUser.Address.City
	}
	if req.Address.State == nil {
		req.Address.State = existingUser.Address.State
	}
	if req.Address.PostalCode == nil {
		req.Address.PostalCode = existingUser.Address.PostalCode
	}
	if req.Address.Country == nil {
		req.Address.Country = existingUser.Address.Country
	}

	// // Merge Vehicle fields
	// if req.Vehicle.Car == nil {
	// 	req.Vehicle.Car = existingUser.Vehicle.Car
	// }
	// if req.Vehicle.Bike == nil {
	// 	req.Vehicle.Bike = existingUser.Vehicle.Bike
	// }
	// Merge Vehicle fields
	if req.Vehicle.Car != nil {
		existingUser.Vehicle.Car = req.Vehicle.Car
	}
	if req.Vehicle.Bike != nil {
		existingUser.Vehicle.Bike = req.Vehicle.Bike
	}
	// Final request = fully merged
	req.Address = existingUser.Address
	req.Vehicle = existingUser.Vehicle

	// Optional: Marshal here for logging/verification/debug (not mandatory)
	addressJSON, err := json.Marshal(req.Address)
	if err != nil {
		return fmt.Errorf("failed to marshal address: %w", err)
	}
	log.Printf("[Service] Merged Address JSON: %s\n", string(addressJSON))

	vehicleJSON, err := json.Marshal(req.Vehicle)
	if err != nil {
		return fmt.Errorf("failed to marshal vehicle: %w", err)
	}
	log.Printf("[Service] Merged Vehicle JSON: %s\n", string(vehicleJSON))

	// Pass updated request to repo (repo will marshal again as per your current design)
	return s.repo.UpdateUserInfo(ctx, req)
}

func (s *service) GetUserProfile(ctx context.Context, req UserLoginRequest) (string, error) {
	log.Println("[Login Service] started")

	user, err := s.repo.FindUserByEmail(ctx, strings.ToLower(req.Email))
	if err != nil {
		log.Println("[FindByEmail Error]:", err)
		return "", fmt.Errorf("user not found")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		log.Println("[Password Mismatch]", err)
		return "", fmt.Errorf("invalid credentials")
	}

	token, err := GenerateToken(fmt.Sprintf("%d", user.ID), user.Email)
	if err != nil {
		log.Println("[Token Generation Error]:", err)
		return "", err
	}
	return token, nil
}

// get profile
func (s *service) GetProfile(ctx context.Context, userID string) (User, error) {
	log.Println("[Service] GetProfile called with userID: ", userID)

	user, addressJSON, vehicleJSON, err := s.repo.FindUserByID(ctx, userID)
	if err != nil {
		log.Println("[GetProfile] Error fetching user:", err)
		return *user, err
	}

	if addressJSON != nil {
		err := json.Unmarshal(addressJSON, &user.Address)
		if err != nil {
			log.Println("[GetProfile] Error unmarshaling address:", err)
			return *user, err
		}
	} else {
		log.Println("[GetProfile] Address is NULL, using zero value")
	}

	if vehicleJSON != nil {
		err := json.Unmarshal(vehicleJSON, &user.Vehicle)
		if err != nil {
			log.Println("[GetProfile] Error unmarshaling vehicle:", err)
			return *user, err
		}
	} else {
		log.Println("[GetProfile] Vehicle is NULL, using zero value")
	}

	return *user, nil
}

// 	return s.repo.FindByUserID(ctx, userID)
// }
