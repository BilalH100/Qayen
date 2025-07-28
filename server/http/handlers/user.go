package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	httpx "kayena/server/http"
	"kayena/server/models"
	"kayena/server/schemas"
	"kayena/server/services"
	"kayena/server/utils"
	"log"
	"net/http"
	"strconv"

	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

func CreateUserHandler(s services.UserService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req schemas.CreateUserRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.RespondWithError(w, fmt.Errorf("invalid request body: %w", err))
			return
		}
		_, err := s.GetUserProfileByEmail(r.Context(), req.Email)
		if err == nil {
			httpx.RespondWithJSON(w, http.StatusConflict, schemas.Message{
				Message: "User already exists",
			})
			return
		} else if !errors.Is(err, pgx.ErrNoRows) {
			httpx.RespondWithError(w, fmt.Errorf("failed to check user existence: %w", err))
			return
		}
		hashedPass, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			httpx.RespondWithError(w, fmt.Errorf("failed to hash password: %w", err))
			return
		}
		user := models.User{
			Name:     req.Name,
			Email:    req.Email,
			Phone:    req.Phone,
			Password: string(hashedPass),
		}

		if err := s.RegisterUser(r.Context(), user); err != nil {
			httpx.RespondWithError(w, fmt.Errorf("failed to register user: %w", err))
			return
		}
		httpx.RespondWithJSON(w, http.StatusCreated, schemas.Message{
			Message: "User registered successfully",
		})
	}
}

func LoginHandler(s services.UserService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req schemas.LoginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.RespondWithError(w, fmt.Errorf("Invalid request body :%w", err))
			return
		}
		user, err := s.GetUserProfileByEmail(r.Context(), req.Email)
		if err == nil {
			log.Println(req.Password)
			log.Println(user.Password)
			if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
				httpx.RespondWithJSON(w, http.StatusBadRequest, schemas.Message{
					Message: "Invalid password",
				})
				return
			}
			token, err := utils.GenerateJWT(*user)
			if err != nil {
				httpx.RespondWithError(w, fmt.Errorf("Error creating jwt: %w", err))
				return
			}
			httpx.RespondWithJSON(w, http.StatusOK, schemas.LoginResponse{
				Token: token,
				User: schemas.GetUserResponse{
					Phone: user.Phone,
					Name:  user.Name,
					Email: user.Email,
					Id:    user.ID,
				},
			})
		} else if errors.Is(err, pgx.ErrNoRows) {
			httpx.RespondWithJSON(w, http.StatusBadRequest, schemas.Message{
				Message: "User not found",
			})
			return
		} else {
			httpx.RespondWithError(w, err)
		}
	}
}

func GetUserProfileHandler(s services.UserService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.URL.Query().Get("id")
		id, err := strconv.ParseInt(idStr, 10, 32)
		if err != nil {
			httpx.RespondWithError(w, fmt.Errorf("Error parsing int:%w", err))
			return
		}
		id32 := int32(id)
		user, err := s.GetUserProfile(r.Context(), id32)
		if err != nil {
			httpx.RespondWithError(w, fmt.Errorf("Error getting user profile: %w", err))
			return
		}
		httpx.RespondWithJSON(w, http.StatusOK, schemas.GetUserResponse{
			Email: user.Email,
			Name:  user.Name,
			Phone: user.Phone,
		})
	}
}
