package userauth

import (
	"context"
	"fmt"
	"log"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type Service interface {
	UserRegister(ctx context.Context, request UserRegisterRequest) (string, error)
	GetUserProfile(ctx context.Context, request UserLoginRequest) (string, error)
	GetProfile(ctx context.Context, userId string) (User, error)
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
		log.Println("[FindByEmail Error]:", err)
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
	// calculate age
	now := time.Now()
	years := now.Year() - user.DOB.Year()
	if now.YearDay() < user.DOB.YearDay() {
		years--
	}
	user.Age = years
	return token, nil
}

// get profile
func (s *service) GetProfile(ctx context.Context, userID string) (User, error) {
	log.Println("[Service] GetProfile called with userID: ", userID)
	return s.repo.FindByUserID(ctx, userID)
}
