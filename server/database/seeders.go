package database

import (
	"context"
	"encoding/json"
	"fmt"
	repository "kayena/server/database/generated"
	"kayena/server/schemas"
	"kayena/server/services"
	"os"
	"time"
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

func SeedPharmacies() error {
	ctx := context.Background()
	f, err := os.ReadFile("database/data/pharmacies.json")
	if err != nil {
		return err
	}

	var pharmacies []schemas.Pharmacy
	err = json.Unmarshal(f, &pharmacies)
	if err != nil {
		return fmt.Errorf("error unmarshal json :%s", err)
	}
	tx, err := DbConn.Begin(ctx)
	if err != nil {
		return err
	}

	defer tx.Rollback(ctx)

	qtx := repository.New(tx)

	for _, phar := range pharmacies {
		geo, err := services.GetGeoCoordinates(phar)
		if err != nil {
			return fmt.Errorf("failed to get coordinates for pharmacy '%s': %w", phar.Name, err)
		}
		fmt.Printf("Coordinates for %s: lat %s lon %s ✅\n", phar.Name, geo.Lat, geo.Long)
		time.Sleep(1 * time.Second)
		
		address, err := services.GetAddressByGeo(*geo)
		if err != nil {
			return fmt.Errorf("failed to get address for pharmacy '%s' (lat: %s, lon: %s): %w", 
				phar.Name, geo.Lat, geo.Long, err)
		}
		fmt.Printf("Address for %s: %s✅ \n", phar.Name, address)
		_, err = qtx.CreatePharmacy(ctx, repository.CreatePharmacyParams{
			Name:      phar.Name,
			City:      phar.City,
			Address:   address,
			Latitude:  geo.Lat,
			Longitude: geo.Long,
		})
		if err != nil {
			return fmt.Errorf("error creating pharmacy record : %s", err)
		}
		time.Sleep(1 * time.Second)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return err
	}
	return nil
}
