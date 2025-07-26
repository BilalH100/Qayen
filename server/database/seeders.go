package database

import (
	"context"
	"encoding/json"
	"fmt"
	repository "kayena/server/database/generated"
	"kayena/server/models"
	"kayena/server/schemas"
	"kayena/server/services"
	"os"
	"sync"
	"time"
)

func SeedMedications(ctx context.Context, s services.MedService) error {
	var (
		medications []models.Medication
		failedMeds  []string
		mu          sync.Mutex
	)

	f, err := os.ReadFile("database/data/medications.json")
	if err != nil {
		return fmt.Errorf("error reading medications.json: %w", err)
	}

	if err := json.Unmarshal(f, &medications); err != nil {
		return fmt.Errorf("error unmarshalling json: %w", err)
	}

	const lim = 10
	sem := make(chan struct{}, lim)
	var wg sync.WaitGroup

	errorsChan := make(chan error, len(medications))

	for _, med := range medications {
		med := med
		wg.Add(1)
		sem <- struct{}{}

		go func() {
			defer wg.Done()
			defer func() { <-sem }()

			description, sideEff, err := services.GetMetaData(med.Speciality, nil)
			if err != nil {
				mu.Lock()
				failedMeds = append(failedMeds, med.Speciality)
				mu.Unlock()
				return
			}

			err = s.CreateMedication(ctx, models.Medication{
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
				Description:      description,
				CommonSd:         sideEff.Common,
				SeriousSd:        sideEff.Serious,
			})
			if err != nil {
				errorsChan <- fmt.Errorf("error creating medication %s: %w", med.Speciality, err)
				return
			}
		}()
	}
	wg.Wait()
	close(errorsChan)

	for err := range errorsChan {
		fmt.Println(err)
	}

	if len(failedMeds) > 0 {
		fmt.Printf("Failed to get metadata for %d medications: %v\n", len(failedMeds), failedMeds)
	}

	return nil
}

func SeedPharmacies() error {
	ctx := context.Background()

	var (
		pharmacies []schemas.Pharmacy
		err        error
		f          []byte
	)
	f, err = os.ReadFile("database/data/pharmacies.json")
	if err != nil {
		return err
	}

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
