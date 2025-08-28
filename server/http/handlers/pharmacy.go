package handlers

import (
	"fmt"
	httpx "kayena/server/http"
	"kayena/server/schemas"
	"kayena/server/services"
	"net/http"
	"strconv"
)

type GetClosestResponse struct {
	Distance int32            `json:"distance"`
	Pharmacy schemas.Pharmacy `json:"pharmacy"`
}

func GetAllPharmaciesHandler(s services.PharmacyService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		res, err := s.GetAllService(r.Context())
		if err != nil {
			fmt.Println("error getting all pharmacies:", err)
			httpx.RespondWithError(w, httpx.ErrInternal)
			return
		}
		pharmacies := make(map[string][]schemas.Pharmacy)
		pharmacies["pharmacies"] = res
		httpx.RespondWithJSON(w, http.StatusOK, pharmacies)
	}
}

func GetClosestPharmacyHandler(s services.PharmacyService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idstr := r.URL.Query().Get("id")
		if idstr == "" {
			fmt.Println("missing medication id")
			httpx.RespondWithError(w, httpx.ErrBadRequest)
			return
		}
		long := r.URL.Query().Get("longitude")
		if long == "" {
			fmt.Println("missing longitude")
			httpx.RespondWithError(w, httpx.ErrBadRequest)
			return
		}
		lat := r.URL.Query().Get("latitude")
		if lat == "" {
			fmt.Println("missing latitude")
			httpx.RespondWithError(w, httpx.ErrBadRequest)
			return
		}
		id, err := strconv.Atoi(idstr)
		if err != nil {
			httpx.RespondWithError(w, httpx.ErrInternal)
			return
		}
		distance, pharmacy, err := s.GetClosestService(r.Context(), schemas.Coordinates{
			Long: long,
			Lat:  lat,
		}, int32(id))
		if err != nil {
			fmt.Println("error getting closest pharmacy:", err)
			httpx.RespondWithError(w, httpx.ErrInternal)
			return
		}
		if distance == nil || pharmacy == nil {
			fmt.Println("no pharmacy found with this medication")
			httpx.RespondWithError(w, fmt.Errorf("no pharmacy found with this medication"))
			return
		}
		httpx.RespondWithJSON(w, http.StatusOK, GetClosestResponse{
			Distance: *distance,
			Pharmacy: *pharmacy,
		})
	}
}
