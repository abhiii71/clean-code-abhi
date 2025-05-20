package userauth

import (
	"context"

	"golang.org/x/crypto/bcrypt"
)

type Service interface {
	Register(ctx context.Context, request UserRegisterRequest) (string, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{
		repo: repo,
	}
}

func (s *service) Register(ctx context.Context, request UserRegisterRequest) (string, error) {
	// business login
	// password hash
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	request.Password = string(hashedPassword)

	// store the user
	userID, err := s.repo.Register(ctx, request)
	if err != nil {
		return "", err
	}

	token, err := GenerateToken(userID, request.Email)
	if err != nil {
		return "", err
	}
	// user exists
	return token, nil
}
