package repository

import (
	"context"
	sqlc "kayena/server/database/generated"
	"kayena/server/models"
)

type UserRepository interface {
	GetById(ctx context.Context, id int32) (*models.User, error)
	Create(ctx context.Context, user *models.User) error
	GetByEmail(ctx context.Context, email string) (*models.User, error)
}

type sqlcUserRepo struct {
	queries *sqlc.Queries
}

func NewUserRepo(db sqlc.DBTX) UserRepository {
	return &sqlcUserRepo{
		queries: sqlc.New(db),
	}
}

func (r *sqlcUserRepo) GetById(ctx context.Context, id int32) (*models.User, error) {
	user, err := r.queries.GetUserByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return &models.User{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
		Phone: user.Phone,
	}, nil
}

func (r *sqlcUserRepo) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	user, err := r.queries.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	return &models.User{
		ID:       user.ID,
		Name:     user.Name,
		Email:    user.Email,
		Phone:    user.Phone,
		Password: user.Password,
	}, nil
}
func (r *sqlcUserRepo) Create(ctx context.Context, user *models.User) error {
	_, err := r.queries.CreateUser(ctx, sqlc.CreateUserParams{
		Email:    user.Email,
		Phone:    user.Phone,
		Name:     user.Name,
		Password: user.Password,
	})
	return err
}
