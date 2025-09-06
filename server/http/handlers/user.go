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

	"github.com/go-chi/chi/v5"
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
			Role:     models.Regular,
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
			httpx.RespondWithError(w, fmt.Errorf("invalid request body :%w", err))
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
				httpx.RespondWithError(w, fmt.Errorf("error creating jwt: %w", err))
				return
			}
			managedStr := strconv.Itoa(int(user.ManagedPharmacyID))
			httpx.RespondWithJSON(w, http.StatusOK, schemas.LoginResponse{
				Token: token,
				User: schemas.GetUserResponse{
					Phone:             user.Phone,
					Name:              user.Name,
					Email:             user.Email,
					Id:                user.ID,
					Role:              user.Role,
					ManagedPharmacyId: managedStr,
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
			httpx.RespondWithError(w, fmt.Errorf("error parsing int:%w", err))
			return
		}
		id32 := int32(id)
		user, err := s.GetUserProfile(r.Context(), id32)
		if err != nil {
			httpx.RespondWithError(w, fmt.Errorf("error getting user profile: %w", err))
			return
		}
		httpx.RespondWithJSON(w, http.StatusOK, schemas.GetUserResponse{
			Email: user.Email,
			Name:  user.Name,
			Phone: user.Phone,
		})
	}
}

func ListAllUsersHandler(s services.UserService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pageStr := r.URL.Query().Get("page")
		limitStr := r.URL.Query().Get("limit")

		page, _ := strconv.Atoi(pageStr)
		limit, _ := strconv.Atoi(limitStr)
		if page < 1 {
			page = 1
		}
		if limit < 1 {
			limit = 10
		}
		offset := (page - 1) * limit

		res, err := s.List(r.Context(), schemas.Options{
			Limit:  int32(limit),
			Offset: int32(offset),
		})
		if err != nil {
			httpx.RespondWithError(w, fmt.Errorf("error getting user list"))
			return
		}
		users := make(map[string][]models.User)
		users["users"] = res
		httpx.RespondWithJSON(w, http.StatusOK, users)
	}
}

func DeleteUserHandler(s services.UserService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idstr := chi.URLParam(r, "id")
		if idstr == "" {
			httpx.RespondWithError(w, fmt.Errorf("user id cannot be empty"))
			return
		}
		id, err := strconv.Atoi(idstr)
		if err != nil {
			httpx.RespondWithError(w, fmt.Errorf("internal server error"))
			return
		}
		err = s.DeleteUser(r.Context(), int32(id))
		if err != nil {
			httpx.RespondWithError(w, fmt.Errorf("internal server error"))
			return
		}
		httpx.RespondWithJSON(w, http.StatusOK, nil)
	}
}

func UpdateUserRole(s services.UserService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idstr := chi.URLParam(r, "id")
		var req schemas.UpdateUserRoleRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.RespondWithError(w, fmt.Errorf("invalid request body :%w", err))
			return
		}
		id, err := strconv.Atoi(idstr)
		if err != nil {
			httpx.RespondWithError(w, fmt.Errorf("internal server error"))
			return
		}
		user, err := s.UpdateUserRole(r.Context(), int32(id), models.Role(req.Role))
		if err != nil {
			httpx.RespondWithError(w, fmt.Errorf("internal server error"))
			return
		}
		httpx.RespondWithJSON(w, http.StatusOK, *user)
	}
}

func UpdateUserManagedPharmacyHandler(s services.UserService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userIdStr := chi.URLParam(r, "id")
		if userIdStr == "" {
			httpx.RespondWithError(w, fmt.Errorf("user id cannot be empty"))
			return
		}
		var req schemas.UpdateUsermanagedPharmacyRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.RespondWithError(w, fmt.Errorf("error decoding request: %w", err))
			return
		}
		fmt.Println("id string", userIdStr)
		userID, err := strconv.Atoi(userIdStr)
		if err != nil {
			httpx.RespondWithError(w, fmt.Errorf("cannot convert user id %w", err))
			return
		}
		// pharmacyID, err := strconv.Atoi(req.PharmacyId)
		// if err != nil {
		// 	httpx.RespondWithError(w, fmt.Errorf("cannot convert pharmacy id %w", err))
		// 	return
		// }
		user, err := s.UpdateUserPharmacy(r.Context(), int32(userID), int32(req.PharmacyId))
		if err != nil {
			httpx.RespondWithError(w, fmt.Errorf("error updating user pharmacy %w", err))
			return
		}
		httpx.RespondWithJSON(w, http.StatusOK, *user)
	}
}
