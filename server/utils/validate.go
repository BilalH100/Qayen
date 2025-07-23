package utils

import "github.com/go-playground/validator/v10"

type CreateUserRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Username string `json:"username" validate:"required,min=3"`
}

var validate = validator.New()

func ValidateStruct(s interface{}) error {
	return validate.Struct(s)
}
