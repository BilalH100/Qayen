package handlers

import (
	"encoding/json"
	"fmt"
	httpx "kayena/server/http"
	"kayena/server/models"
	"kayena/server/schemas"
	"kayena/server/services"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type GetClosestResponse struct {
	Distance float64          `json:"distance"`
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
		pharmacies := make(map[string][]models.Pharmacy)
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

func GetPharmacyDetailsHandler(s services.PharmacyService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		if idStr == "" {
			httpx.RespondWithError(w, fmt.Errorf("id cannot be empty"))
			return
		}
		id, err := strconv.Atoi(idStr)
		if err != nil {
			httpx.RespondWithError(w, fmt.Errorf("error parsing id string"))
			return
		}
		res, err := s.GetByIdService(r.Context(), int32(id))
		if err != nil {
			httpx.RespondWithError(w, fmt.Errorf("internal server error"))
			return
		}
		pharmacy := make(map[string]schemas.Pharmacy)
		pharmacy["pharmacy"] = *res
		httpx.RespondWithJSON(w, http.StatusOK, pharmacy)
	}
}

// ListPharmacyStockHandler returns everything a pharmacy currently has on hand.
func ListPharmacyStockHandler(s services.PharmacyService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			httpx.RespondWithError(w, fmt.Errorf("invalid pharmacy id"))
			return
		}
		items, err := s.ListStockService(r.Context(), int32(id))
		if err != nil {
			fmt.Println("error listing stock:", err)
			httpx.RespondWithError(w, httpx.ErrInternal)
			return
		}
		httpx.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
			"success": true,
			"data":    items,
		})
	}
}

type upsertStockRequest struct {
	MedicationID int32 `json:"medication_id"`
	Quantity     int32 `json:"quantity"`
}

// UpsertPharmacyStockHandler adds a medication to a pharmacy's stock, or updates the quantity if it's already there.
func UpsertPharmacyStockHandler(s services.PharmacyService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			httpx.RespondWithError(w, fmt.Errorf("invalid pharmacy id"))
			return
		}

		var req upsertStockRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.RespondWithError(w, fmt.Errorf("invalid request body"))
			return
		}
		if req.MedicationID <= 0 {
			httpx.RespondWithError(w, fmt.Errorf("medication_id is required"))
			return
		}
		if req.Quantity < 0 {
			httpx.RespondWithError(w, fmt.Errorf("quantity cannot be negative"))
			return
		}

		item, err := s.UpsertStockService(r.Context(), int32(id), req.MedicationID, req.Quantity)
		if err != nil {
			fmt.Println("error upserting stock:", err)
			httpx.RespondWithError(w, httpx.ErrInternal)
			return
		}
		httpx.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
			"success": true,
			"data":    item,
		})
	}
}

// DeletePharmacyStockHandler removes one medication from a pharmacy's stock.
func DeletePharmacyStockHandler(s services.PharmacyService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			httpx.RespondWithError(w, fmt.Errorf("invalid pharmacy id"))
			return
		}
		medIdStr := chi.URLParam(r, "medId")
		medId, err := strconv.Atoi(medIdStr)
		if err != nil {
			httpx.RespondWithError(w, fmt.Errorf("invalid medication id"))
			return
		}
		if err := s.DeleteStockService(r.Context(), int32(id), int32(medId)); err != nil {
			fmt.Println("error deleting stock:", err)
			httpx.RespondWithError(w, httpx.ErrInternal)
			return
		}
		httpx.RespondWithJSON(w, http.StatusOK, map[string]interface{}{"success": true})
	}
}
