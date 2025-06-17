package userauth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type Service interface {
	UserRegister(ctx context.Context, request UserRegisterRequest) (string, error)
	GetUserProfile(ctx context.Context, request UserLoginRequest) (string, error)
	GetProfile(ctx context.Context, userId string) (User, error)
	UpdateUserInfo(ctx context.Context, req UserInformationRequest) error
	SaveUserPDF(ctx context.Context, email string, file *multipart.FileHeader) error
	// GetUserInfo(ctx context.Context, request UserInformationRequest) error
}

type service struct {
	repo      Repository
	uploadDir string
}

func NewService(repo Repository) Service {
	return &service{
		repo: repo,
	}
}

func (s *service) UserRegister(ctx context.Context, request UserRegisterRequest) (string, error) {
	// business logic
	// password hash
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Println("[hashed password Error]:", err)
		return "", err
	}

	_, err = s.repo.FindUserByEmail(ctx, request.Email)
	if err == nil {
		return "", fmt.Errorf("email already regestered")
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

func (s *service) UpdateUserInfo(ctx context.Context, req UserInformationRequest) error {
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

	if req.Address.City != nil {
		existingUser.Address.City = req.Address.City
	}
	if req.Address.State != nil {
		existingUser.Address.State = req.Address.State
	}
	if req.Address.PostalCode != nil {
		existingUser.Address.PostalCode = req.Address.PostalCode
	}
	if req.Address.Country != nil {
		existingUser.Address.Country = req.Address.Country
	}

	if req.Vehicle.Car != nil {
		existingUser.Vehicle.Car = req.Vehicle.Car
	}
	if req.Vehicle.Bike != nil {
		existingUser.Vehicle.Bike = req.Vehicle.Bike
	}

	// Marshal merged data
	addressJSON, err := json.Marshal(existingUser.Address)
	if err != nil {
		return fmt.Errorf("failed to marshal address: %w", err)
	}

	vehicleJSON, err := json.Marshal(existingUser.Vehicle)
	if err != nil {
		return fmt.Errorf("failed to marshal vehicle: %w", err)
	}

	log.Println("Final addressJSON:", string(addressJSON))
	log.Println("Final vehicleJSON:", string(vehicleJSON))

	// Save to DB
	return s.repo.UpdateUserInfo(ctx, req, addressJSON, vehicleJSON)
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

// get profilefunc calculateAge(dob time.Time) int {
func calculateAge(dob time.Time) int {
	now := time.Now()
	years := now.Year() - dob.Year()
	if now.YearDay() < dob.YearDay() {
		years--
	}
	return years
}

func (s *service) GetProfile(ctx context.Context, userID string) (User, error) {
	log.Println("[Service] GetProfile called with userID:", userID)

	user, addressJSON, vehicleJSON, err := s.repo.FindUserByID(ctx, userID)
	if err != nil {
		log.Println("[GetProfile] Error fetching user:", err)
		return User{}, err
	}

	// Calculate age from DOB
	user.Age = calculateAge(user.DOB)

	// Unmarshal Address if present
	if addressJSON != nil {
		if err := json.Unmarshal(addressJSON, &user.Address); err != nil {
			log.Println("[GetProfile] Error unmarshaling address:", err)
			return User{}, err
		}
	} else {
		log.Println("[GetProfile] Address is NULL, using zero value")
	}

	// Unmarshal Vehicle if present
	if vehicleJSON != nil {
		if err := json.Unmarshal(vehicleJSON, &user.Vehicle); err != nil {
			log.Println("[GetProfile] Error unmarshaling vehicle:", err)
			return User{}, err
		}
	} else {
		log.Println("[GetProfile] Vehicle is NULL, using zero value")
	}

	return *user, nil
}

func sanitizeEmail(email string) string {
	email = strings.ReplaceAll(email, "@", "_at_")
	email = strings.ReplaceAll(email, ".", "_dot_")
	return email
}

func (s *service) saveFile(file *multipart.FileHeader, email string) error {
	safeEmail := sanitizeEmail(email)
	filePath := filepath.Join(s.uploadDir, safeEmail+".pdf")

	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	return err
}
func (s *service) SaveUserPDF(ctx context.Context, email string, file *multipart.FileHeader) error {
	return s.saveFile(file, email)
}

// func (s *service) UpdateUserInfo(ctx context.Context, request UserInformationRequest) error {
// 	// Try to update first
// 	err := s.repo.UpdateUserInfo(ctx, request)
// 	return err
// }

// // User information update

// // Helper functions to check empty structs
// func isEmptyAddress(addr Address) bool {
// 	return addr.City == nil && addr.State == nil && addr.PostalCode == nil && addr.Country == nil
// }

// func isEmptyVehicle(v Vehicle) bool {
// 	// Customize depending on what empty means
// 	return (v.Car == nil && v.Bike == nil)
// }
