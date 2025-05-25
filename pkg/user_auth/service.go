package userauth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"golang.org/x/crypto/bcrypt"
)

type Service interface {
	UserRegister(ctx context.Context, request UserRegisterRequest) (string, error)
	GetUserProfile(ctx context.Context, request UserLoginRequest) (string, error)
	GetProfile(ctx context.Context, userId string) (User, error)
	UserInformation(ctx context.Context, request UserInformationRequest) error
	UpdateUserInformation(ctx context.Context, req UserInformationRequest) error
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

	_, err = s.repo.FindByEmail(ctx, request.Email)
	if err == nil {
		// log.Println("[FindByEmail Error]:", err)
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

//	func (s *service) UserInformation(ctx context.Context, request UserInformationRequest) error {
//		// validate the request
//		if request.ID == 0 {
//			return fmt.Errorf("invalid user id")
//		}
//		err := s.repo.InsertUserInformation(ctx, request)
//		if err != nil {
//			log.Println("[UserInformation] Insert failed:", err)
//			return err
//		}
//		return nil
//	}
var ErrNoRowsAffected = errors.New("no rows affected")

func (s *service) UserInformation(ctx context.Context, req UserInformationRequest) error {
	// Try to update first
	err := s.repo.UpsertUserInformation(ctx, req)
	return err
}

// User information update
func (s *service) UpdateUserInformation(ctx context.Context, req UserInformationRequest) error {
	return s.repo.UpsertUserInformation(ctx, req)
}
func (s *service) GetUserProfile(ctx context.Context, req UserLoginRequest) (string, error) {
	log.Println("[Login Service] started")

	user, err := s.repo.FindByEmail(ctx, req.Email)
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

	user, addressJSON, vehicleJSON, err := s.repo.FindByUserID(ctx, userID)
	if err != nil {
		log.Println("[GetProfile] Error fetching user:", err)
		return User{}, err
	}

	if addressJSON != nil {
		err := json.Unmarshal(addressJSON, &user.Address)
		if err != nil {
			log.Println("[GetProfile] Error unmarshaling address:", err)
			return User{}, err
		}
	} else {
		log.Println("[GetProfile] Address is NULL, using zero value")
	}

	if vehicleJSON != nil {
		err := json.Unmarshal(vehicleJSON, &user.Vehicle)
		if err != nil {
			log.Println("[GetProfile] Error unmarshaling vehicle:", err)
			return User{}, err
		}
	} else {
		log.Println("[GetProfile] Vehicle is NULL, using zero value")
	}

	return user, nil
}

// 	return s.repo.FindByUserID(ctx, userID)
// }
