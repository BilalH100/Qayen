package services

import (
	"context"
	"kayena/server/models"
	"kayena/server/repository"
	"kayena/server/schemas"
)

type UserService interface {
	UpdateProfile(ctx context.Context, id int32, name, phone string) (*models.User, error)
	GoogleLogin(ctx context.Context, subject, email, name string) (*models.User, error)
	GetUserProfile(ctx context.Context, id int32) (*models.User, error)
	GetUserProfileByEmail(ctx context.Context, email string) (*models.User, error)
	RegisterUser(ctx context.Context, user models.User) error
	List(ctx context.Context, options schemas.Options) ([]models.User, error)
	CountByRole(ctx context.Context) (map[models.Role]int64, error)
	DeleteUser(ctx context.Context, id int32) error
	UpdateUserRole(ctx context.Context, id int32, role models.Role) (*models.User, error)
	UpdateUserPharmacy(ctx context.Context, userID int32, pharmacyId int32) (*models.User, error)
}

func (s *userService) UpdateProfile(ctx context.Context, id int32, name, phone string) (*models.User, error) {
	return s.userRepo.UpdateProfile(ctx, id, name, phone)
}

func (s *userService) GoogleLogin(ctx context.Context, subject, email, name string) (*models.User, error) {
	return s.userRepo.GoogleLogin(ctx, subject, email, name)
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

func (s *userService) List(ctx context.Context, options schemas.Options) ([]models.User, error) {
	return s.userRepo.List(ctx, options)
}

func (s *userService) CountByRole(ctx context.Context) (map[models.Role]int64, error) {
	return s.userRepo.CountByRole(ctx)
}

func (s *userService) DeleteUser(ctx context.Context, id int32) error {
	return s.userRepo.Delete(ctx, id)
}

func (s *userService) UpdateUserRole(ctx context.Context, id int32, role models.Role) (*models.User, error) {
	return s.userRepo.UpdateRole(ctx, id, role)
}

func (s *userService) UpdateUserPharmacy(ctx context.Context, userID int32, pharmacyId int32) (*models.User, error) {
	return s.userRepo.UpdateManagedPharmacy(ctx, userID, pharmacyId)
}
