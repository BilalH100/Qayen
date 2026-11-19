package repository

import (
	"context"
	"kayena/server/models"
)

func (r *sqlcUserRepo) UpdateProfile(ctx context.Context, id int32, name, phone string) (*models.User, error) {
	user := &models.User{}
	err := r.db.QueryRow(ctx, `UPDATE users SET name=$2,phone=$3 WHERE id=$1 RETURNING id,email,COALESCE(name,''),COALESCE(phone,''),user_role,COALESCE(managed_pharmacy_id,0)`, id, name, phone).Scan(&user.ID, &user.Email, &user.Name, &user.Phone, &user.Role, &user.ManagedPharmacyID)
	if err != nil {
		return nil, err
	}
	return user, nil
}
