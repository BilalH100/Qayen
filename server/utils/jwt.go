package utils

import (
	"kayena/server/models"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func GenerateJWT(user models.User) (string, error) {
	secret := os.Getenv("JWT_SECRECT")
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":  user.ID,
		"name": user.Name,
		"exp":  jwt.NewNumericDate(time.Now().Add(time.Hour * 72)),
	})
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return err.Error(), err
	}
	return tokenString, nil
}
