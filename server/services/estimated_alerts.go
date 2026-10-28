package services

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	sqlc "kayena/server/database/generated"
)

func (s *AlertService) isRecordedTestAlert(ctx context.Context, id int32) (bool, error) {
	var exists bool
	if err := s.db.QueryRow(ctx, `SELECT to_regclass('catalog_test_alerts') IS NOT NULL`).Scan(&exists); err != nil || !exists {
		return false, err
	}
	var isTest bool
	err := s.db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM catalog_test_alerts WHERE alert_id=$1)`, id).Scan(&isTest)
	return isTest, err
}

func (s *AlertService) alertStockTable() string {
	if s.testApproximate {
		return "alert_stock"
	}
	return "stock"
}
func (s *AlertService) alertLocationTable() string {
	if s.testApproximate {
		return "alert_pharmacy_locations"
	}
	return "pharmacies"
}

func (s *AlertService) decrementAlertStock(ctx context.Context, tx pgx.Tx, arg sqlc.DecrementStockQuantityParams) (int32, error) {
	if s.testApproximate {
		var id int32
		if err := tx.QueryRow(ctx, `SELECT id FROM alert_stock WHERE pharmacy_id=$1 AND medication_id=$2`, arg.PharmacyID, arg.MedicationID).Scan(&id); err != nil {
			return 0, err
		}
		if id < 0 {
			var remaining int32
			err := tx.QueryRow(ctx, `UPDATE catalog_test_stock SET quantity=quantity-$3,updated_at=now() WHERE pharmacy_id=$1 AND medication_id=$2 AND quantity >= $3 AND current_setting('qayen.test_alerts',true)='on' AND EXISTS(SELECT 1 FROM catalog_test_settings WHERE enabled) RETURNING quantity`, arg.PharmacyID, arg.MedicationID, arg.Quantity).Scan(&remaining)
			return remaining, err
		}
	}
	row, err := s.queries.WithTx(tx).DecrementStockQuantity(ctx, arg)
	return row.Quantity, err
}

func (s *AlertService) testAvailability(ctx context.Context, id int32, lat, lon float64) ([]PharmacyAvailability, error) {
	rows, err := s.db.Query(ctx, `SELECT p.id,p.name,p.address,p.phone,p.latitude,p.longitude,
 6371*acos(LEAST(1.0,GREATEST(-1.0,cos(radians($2::float8))*cos(radians(p.latitude))*cos(radians(p.longitude)-radians($3::float8))+sin(radians($2::float8))*sin(radians(p.latitude))))),
 pr.response_type,COALESCE(pr.substitute_brand,''),COALESCE(pr.substitute_notes,''),COALESCE(pr.response_time_seconds,0),COALESCE(st.quantity,0),
 (b.latitude IS NULL AND b.longitude IS NULL),COALESCE(st.id<0,false)
 FROM pharmacist_responses pr JOIN alert_pharmacy_locations p ON p.id=pr.pharmacy_id
 JOIN pharmacies b ON b.id=p.id JOIN medication_alerts ma ON ma.id=pr.alert_id
 LEFT JOIN alert_stock st ON st.pharmacy_id=p.id AND st.medication_id=ma.medication_id
 WHERE pr.alert_id=$1 AND pr.expires_at>now() ORDER BY 7,p.id`, id, lat, lon)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []PharmacyAvailability{}
	for rows.Next() {
		var p PharmacyAvailability
		if err = rows.Scan(&p.PharmacyID, &p.PharmacyName, &p.PharmacyAddress, &p.PharmacyPhone, &p.Latitude, &p.Longitude, &p.DistanceKM, &p.ResponseType, &p.SubstituteBrand, &p.SubstituteNotes, &p.ResponseTime, &p.AvailableQuantity, &p.EstimatedLocation, &p.TestInventory); err != nil {
			return nil, err
		}
		result = append(result, p)
	}
	return result, rows.Err()
}

func (s *AlertService) testDashboard(ctx context.Context, id int32) ([]sqlc.GetPharmacyDashboardAlertsRow, error) {
	rows, err := s.db.Query(ctx, `SELECT ma.id,ma.created_at,ma.expires_at,m.speciality,m.active_substance,u.name,
 6371*acos(LEAST(1.0,GREATEST(-1.0,cos(radians(ma.customer_latitude))*cos(radians(p.latitude))*cos(radians(p.longitude)-radians(ma.customer_longitude))+sin(radians(ma.customer_latitude))*sin(radians(p.latitude))))),
 (pr.id IS NOT NULL)
 FROM medication_alerts ma JOIN medications m ON m.id=ma.medication_id JOIN users u ON u.id=ma.customer_id
 JOIN pharmacy_alert_notifications pan ON pan.alert_id=ma.id JOIN alert_pharmacy_locations p ON p.id=pan.pharmacy_id
 LEFT JOIN pharmacist_responses pr ON pr.alert_id=ma.id AND pr.pharmacy_id=$1
 WHERE pan.pharmacy_id=$1 AND p.latitude IS NOT NULL AND p.longitude IS NOT NULL AND ma.status='pending' AND ma.expires_at>now() ORDER BY ma.created_at DESC`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []sqlc.GetPharmacyDashboardAlertsRow
	for rows.Next() {
		var r sqlc.GetPharmacyDashboardAlertsRow
		if err = rows.Scan(&r.ID, &r.CreatedAt, &r.ExpiresAt, &r.MedicationName, &r.ActiveSubstance, &r.CustomerName, &r.CustomerDistanceKm, &r.AlreadyResponded); err != nil {
			return nil, err
		}
		result = append(result, r)
	}
	return result, rows.Err()
}

func (s *AlertService) dashboardRows(ctx context.Context, id int32) ([]sqlc.GetPharmacyDashboardAlertsRow, error) {
	if s.testApproximate {
		return s.testDashboard(ctx, id)
	}
	return s.queries.GetPharmacyDashboardAlerts(ctx, pgtype.Int4{Int32: id, Valid: true})
}
