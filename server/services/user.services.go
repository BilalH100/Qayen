package services

import (
	"context"
	"kayena/server/models"
	"kayena/server/repository"
)

type UserService interface {
	GetUserProfile(ctx context.Context, id int32) (*models.User, error)
	GetUserProfileByEmail(ctx context.Context, name string) (*models.User, error)
	RegisterUser(ctx context.Context, user models.User) error
}

type userService struct {
	userRepo repository.UserRepository
}

func NewUserService(userRepo repository.UserRepository) UserService {
	return &userService{userRepo: userRepo}
}

func (s *userService) GetUserProfile(ctx context.Context, id int32) (*models.User, error) {
	return s.userRepo.GetById(ctx, id)
}

func (s *userService) GetUserProfileByEmail(ctx context.Context, email string) (*models.User, error) {
	return s.userRepo.GetByEmail(ctx, email)
}

func (s *userService) RegisterUser(ctx context.Context, user models.User) error {
	return s.userRepo.Create(ctx, &user)
}
