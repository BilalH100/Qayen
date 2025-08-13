package handlers

import (
	"encoding/json"
	"fmt"
	httpx "kayena/server/http"
	"kayena/server/models"
	"kayena/server/schemas"
	"kayena/server/services"
	"log"
	"net/http"
	"strconv"
)

func GetMedicationByIdHandler(s services.MedService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idstr := r.PathValue("id")
		if idstr == "" {
			httpx.RespondWithError(w, fmt.Errorf("Error :Medication Id is required"))
			return
		}
		id, err := strconv.ParseInt(idstr, 10, 32)
		if err != nil {
			httpx.RespondWithError(w, err)
			return
		}
		id32 := int32(id)
		med, err := s.GetMedicationById(r.Context(), id32)
		if err != nil {
			httpx.RespondWithError(w, err)
			return
		}
		httpx.RespondWithJSON(w, http.StatusOK, med)
	}
}

func GetMedicationByCodeHandler(s services.MedService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		codestr := r.PathValue("code")
		if codestr == "" {
			httpx.RespondWithError(w, fmt.Errorf("Error : Medication code is required"))
			return
		}
		med, err := s.GetMedicationByCode(r.Context(), codestr)
		if err != nil {
			httpx.RespondWithError(w, err)
			return
		}
		httpx.RespondWithJSON(w, http.StatusOK, med)
	}
}

func CreateMedicationHandler(s services.MedService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req schemas.CreateMedicationRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.RespondWithError(w, fmt.Errorf("Invalid body request :%w", err))
			return
		}
		log.Println(req)
		err := s.CreateMedication(r.Context(), models.Medication{
			Status:           req.Status,
			CommercialStatus: req.CommercialStatus,
			Speciality:       req.Speciality,
			Dosage:           req.Dosage,
			Form:             req.Form,
			Presentation:     req.Presentation,
			Pp:               req.Pp,
			ActiveSubstance:  req.ActiveSubstance,
			TherapeuticClass: req.TherapeuticClass,
			Epi:              req.Epi,
			Ppv:              req.Ppv,
			Ph:               req.Ph,
			Pfht:             req.Pfht,
			Code:             req.Code,
			Tva:              req.Tva,
			Description:      req.Description,
			CommonSd:         req.CommonSd,
			SeriousSd:        req.SeriousSd,
			GeneralInfo:      req.GeneralInfo,
		})
		if err != nil {
			httpx.RespondWithError(w, err)
			return
		}
		httpx.RespondWithJSON(w, http.StatusOK, schemas.Message{
			Message: "Medication created successfully",
		})
	}
}

func DeleteMedicationHandler(s services.MedService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idstr := r.URL.Query().Get("id")
		if idstr == "" {
			httpx.RespondWithError(w, fmt.Errorf("Error: Medication Id is required"))
			return
		}
		id, err := strconv.ParseInt(idstr, 10, 32)
		if err != nil {
			httpx.RespondWithError(w, err)
			return
		}
		id32 := int32(id)
		err = s.DeleteMedication(r.Context(), id32)
		if err != nil {
			httpx.RespondWithError(w, err)
			return
		}
		httpx.RespondWithJSON(w, http.StatusOK, schemas.Message{
			Message: "Medication deleted successfully",
		})
	}
}

func UpdateMedicationHandler(s services.MedService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req schemas.CreateMedicationRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.RespondWithError(w, fmt.Errorf("Invalid request body :%w", err))
			return
		}
		s.UpdateMedication(r.Context(), models.Medication{})
	}
}

func GetMedicationsHandler(s services.MedService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			limit  int64
			offset int64
			err    error
		)

		limitStr := r.URL.Query().Get("limit")
		offsetStr := r.URL.Query().Get("offset")
		if limitStr == "" {
			limit = 100
		} else {
			limit, err = strconv.ParseInt(limitStr, 10, 32)
			if err != nil {
				httpx.RespondWithError(w, err)
				return
			}
		}
		if offsetStr == "" {
			offset = 10
		} else {
			offset, err = strconv.ParseInt(offsetStr, 10, 32)
			if err != nil {
				httpx.RespondWithError(w, err)
				return
			}
		}
		meds, err := s.GetAllMedications(r.Context(), schemas.Options{
			Limit:  int32(limit),
			Offset: int32(offset),
		})

		if err != nil {
			httpx.RespondWithError(w, err)
			return
		}
		log.Println(len(meds))
		httpx.RespondWithJSON(w, http.StatusOK, meds)
	}
}

func GetMedsCategories(s services.MedService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cat, err := s.GetAllCategories(r.Context())
		if err != nil {
			httpx.RespondWithError(w, err)
			return
		}
		httpx.RespondWithJSON(w, http.StatusOK, cat)
	}
}

func SearchMedicationsHandler(s services.MedService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query().Get("q")
		if query == "" {
			httpx.RespondWithError(w, fmt.Errorf("search query is required"))
			return
		}

		medications, err := s.SearchMedications(r.Context(), query)
		if err != nil {
			fmt.Println(err)
			httpx.RespondWithError(w, err)
			return
		}

		httpx.RespondWithJSON(w, http.StatusOK, medications)
	}
}
