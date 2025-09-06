package repository

import (
	"context"
	sqlc "kayena/server/database/generated"
	"kayena/server/models"
	"kayena/server/schemas"
	"kayena/server/utils"
)

type UserRepository interface {
	GetById(ctx context.Context, id int32) (*models.User, error)
	Create(ctx context.Context, user *models.User) error
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	List(ctx context.Context, options schemas.Options) ([]models.User, error)
	Delete(ctx context.Context, id int32) error
	UpdateRole(ctx context.Context, id int32, role models.Role) (*models.User, error)
	UpdateManagedPharmacy(ctx context.Context, userID int32, pharmacyID int32) (*models.User, error)
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
		return nil, wrap(err, "")
	}

	return &models.User{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
		Phone: user.Phone,
		Role:  models.Role(user.UserRole),
	}, nil
}

func (r *sqlcUserRepo) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	user, err := r.queries.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, wrap(err, "")
	}

	return &models.User{
		ID:                user.ID,
		Name:              user.Name,
		Email:             user.Email,
		Phone:             user.Phone,
		Password:          user.Password,
		Role:              models.Role(user.UserRole),
		ManagedPharmacyID: user.ManagedPharmacyID.Int32,
	}, nil
}
func (r *sqlcUserRepo) Create(ctx context.Context, user *models.User) error {
	_, err := r.queries.CreateUser(ctx, sqlc.CreateUserParams{
		Email:    user.Email,
		Phone:    user.Phone,
		Name:     user.Name,
		Password: user.Password,
		UserRole: sqlc.UserRole(user.Role),
	})
	return err
}

func (r *sqlcUserRepo) List(ctx context.Context, options schemas.Options) ([]models.User, error) {
	var users []models.User
	res, err := r.queries.ListUsers(ctx, sqlc.ListUsersParams{
		Limit:  options.Limit,
		Offset: options.Offset,
	})

	if err != nil {
		return nil, wrap(err, "")
	}
	for _, user := range res {
		users = append(users, models.User{
			ID:    user.ID,
			Name:  user.Name,
			Email: user.Email,
			Phone: user.Phone,
			Role:  models.Role(user.UserRole),
		})
	}
	return users, nil
}

func (r *sqlcUserRepo) Delete(ctx context.Context, id int32) error {
	err := r.queries.DeleteUser(ctx, id)
	if err != nil {
		return wrap(err, "")
	}
	return nil
}

func (r *sqlcUserRepo) UpdateRole(ctx context.Context, id int32, role models.Role) (*models.User, error) {
	res, err := r.queries.UpdateUserRole(ctx, sqlc.UpdateUserRoleParams{
		ID:       id,
		UserRole: sqlc.UserRole(role),
	})
	if err != nil {
		return nil, wrap(err, "")
	}
	return &models.User{
		ID:    res.ID,
		Phone: res.Phone,
		Role:  models.Role(res.UserRole),
		Email: res.Email,
		Name:  res.Name,
	}, nil
}

func (r *sqlcUserRepo) UpdateManagedPharmacy(ctx context.Context, userID int32, pharmacyID int32) (*models.User, error) {
	res, err := r.queries.UpdateUserManagedPharmacy(ctx, sqlc.UpdateUserManagedPharmacyParams{
		ID:                userID,
		ManagedPharmacyID: utils.Int32ToPgInt4(pharmacyID),
	})
	if err != nil {
		return nil, wrap(err, "")
	}
	return &models.User{
		ID:                res.ID,
		Name:              res.Name,
		Email:             res.Email,
		Phone:             res.Phone,
		Role:              models.Role(res.UserRole),
		ManagedPharmacyID: res.ManagedPharmacyID.Int32,
	}, nil
}
