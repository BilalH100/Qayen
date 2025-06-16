package database

import (
	"context"
	"encoding/json"
	"fmt"
	repository "kayena/server/database/generated"
	"kayena/server/schemas"
	"os"
)

func SeedMedications() error {
	f, err := os.ReadFile("database/data/medications.json")
	if err != nil {
		return err
	}

	var medications []schemas.Medication
	err = json.Unmarshal(f, &medications)
	if err != nil {
		return fmt.Errorf("error unmarshl json :%s", err)
	}

	ctx := context.Background()
	tx, err := DbConn.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	qtx := repository.New(tx)

	for _, med := range medications {
		_, err := qtx.CreateMedication(ctx, repository.CreateMedicationParams{
			Status:           med.Status,
			CommercialStatus: med.CommercialStatus,
			Speciality:       med.Speciality,
			Dosage:           med.Dosage,
			Form:             med.Form,
			Presentation:     med.Presentation,
			Pp:               med.Pp,
			ActiveSubstance:  med.ActiveSubstance,
			TherapeuticClass: med.TherapeuticClass,
			Epi:              med.Epi,
			Ppv:              med.Ppv,
			Ph:               med.Ph,
			Code:             med.Code,
			Pfht:             med.Pfht,
			Tva:              med.Tva,
		})
		if err != nil {
			return fmt.Errorf("error creating med record: %s", err)
		}
	}
	if err = tx.Commit(ctx); err != nil {
		return err
	}
	return nil
}
