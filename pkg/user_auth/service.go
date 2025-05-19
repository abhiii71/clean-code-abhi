package userauth

import "context"

type Service interface {
	Register(ctx context.Context, request UserRegisterRequest) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{
		repo: repo,
	}
}

func (s *service) Register(ctx context.Context, request UserRegisterRequest) error {
	// business login
	// password hash
	// user exists
	return s.repo.Register(ctx, request)
}
