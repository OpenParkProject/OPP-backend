package dao

import (
	"OPP/backend/api"
	"OPP/backend/db"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

var (
	ErrFineNotFound    = errors.New("fine not found")
	ErrFineAlreadyPaid = errors.New("fine already paid")
)

type FineDao struct {
	db db.DB
}

func NewFineDao() *FineDao {
	return &FineDao{
		db: *db.GetDB(),
	}
}

func (d *FineDao) GetFines(c context.Context, limit *int, offset *int) []api.FineResponse {
	query := "SELECT id, plate, amount, date, paid, zone_id FROM fines LIMIT $1 OFFSET $2"
	params := []any{20, 0}
	if limit != nil {
		params[0] = *limit
	}
	if offset != nil {
		params[1] = *offset
	}

	fines := []api.FineResponse{}
	rows, err := d.db.Query(c, query, params...)
	if err != nil {
		fmt.Printf("db error: %v\n", err.Error())
		return fines
	}
	defer rows.Close()

	for rows.Next() {
		var fine api.FineResponse
		if err := rows.Scan(&fine.Id, &fine.Plate, &fine.Amount, &fine.Date, &fine.Paid, &fine.ZoneId); err != nil {
			fmt.Printf("row scan error: %v\n", err.Error())
			continue
		}
		fines = append(fines, fine)
	}

	return fines
}

func (d *FineDao) GetCarFines(c context.Context, plate string) []api.FineResponse {
	query := "SELECT id, plate, amount, date, paid, zone_id FROM fines WHERE plate = $1"
	rows, err := d.db.Query(c, query, plate)
	if err != nil {
		fmt.Printf("db error: %v\n", err.Error())
		return []api.FineResponse{}
	}
	defer rows.Close()

	fines := []api.FineResponse{}
	for rows.Next() {
		var fine api.FineResponse
		if err := rows.Scan(&fine.Id, &fine.Plate, &fine.Amount, &fine.Date, &fine.Paid, &fine.ZoneId); err != nil {
			fmt.Printf("row scan error: %v\n", err.Error())
			continue
		}
		fines = append(fines, fine)
	}

	return fines
}

func (d *FineDao) AddCarFine(c context.Context, plate string, fine api.FineRequest) (*api.FineResponse, error) {
	carQuery := "SELECT * FROM cars WHERE plate = $1"
	carRows, err := d.db.Query(c, carQuery, plate)
	if err != nil {
		return nil, fmt.Errorf("failed to check car: %w", err)
	}
	defer carRows.Close()

	if !carRows.Next() {
		return nil, ErrCarNotFound
	}

	res, err := NewZoneDao().ZoneExists(c, fine.ZoneId)
	if err != nil {
		return nil, fmt.Errorf("failed to check zone existence: %w", err)
	}
	if !res {
		return nil, fmt.Errorf("zone with ID %d does not exist", fine.ZoneId)
	}

	query := "INSERT INTO fines (plate, amount, date, paid, zone_id) VALUES ($1, $2, $3, FALSE, $4) RETURNING id"
	currentDate := time.Now()
	var lastId int64
	err = d.db.QueryRow(c, query, plate, fine.Amount, currentDate, fine.ZoneId).Scan(&lastId)
	if err != nil {
		return nil, fmt.Errorf("failed to add fine: %w", err)
	}

	return &api.FineResponse{
		Id:     lastId,
		Plate:  plate,
		Amount: fine.Amount,
		Date:   currentDate,
		Paid:   false,
		ZoneId: fine.ZoneId,
	}, nil
}

func (d *FineDao) GetUserFines(c context.Context, username string) ([]api.FineResponse, error) {
	query := "SELECT f.id, f.plate, f.amount, f.date, f.paid, f.zone_id FROM fines f JOIN cars c ON f.plate = c.plate WHERE c.user_id = $1"
	rows, err := d.db.Query(c, query, username)
	if err != nil {
		return nil, fmt.Errorf("failed to get user fines: %w", err)
	}
	defer rows.Close()

	fines := []api.FineResponse{}
	for rows.Next() {
		var fine api.FineResponse
		if err := rows.Scan(&fine.Id, &fine.Plate, &fine.Amount, &fine.Date, &fine.Paid, &fine.ZoneId); err != nil {
			return nil, fmt.Errorf("row scan error: %w", err)
		}
		fines = append(fines, fine)
	}

	return fines, nil
}

func (d *FineDao) GetFineById(c context.Context, id int64) (*api.FineResponse, error) {
	query := "SELECT id, plate, amount, date, paid, zone_id FROM fines WHERE id = $1"
	row := d.db.QueryRow(c, query, id)

	var fine api.FineResponse
	if err := row.Scan(&fine.Id, &fine.Plate, &fine.Amount, &fine.Date, &fine.Paid, &fine.ZoneId); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrFineNotFound
		}
		return nil, fmt.Errorf("failed to get fine by id: %w", err)
	}

	return &fine, nil
}

func (d *FineDao) DeleteFines(c context.Context) error {
	query := "DELETE FROM fines"
	_, err := d.db.Exec(c, query)
	if err != nil {
		return fmt.Errorf("failed to delete fines: %w", err)
	}
	return nil
}

func (d *FineDao) DeleteFineById(c context.Context, id int64) error {
	query := "DELETE FROM fines WHERE id = $1"
	result, err := d.db.Exec(c, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete fine: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrFineNotFound
	}

	return nil
}

func (d *FineDao) PayFine(c context.Context, id int64) error {
	checkQuery := "SELECT paid FROM fines WHERE id = $1"
	var isPaid bool
	err := d.db.QueryRow(c, checkQuery, id).Scan(&isPaid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrFineNotFound
		}
		return fmt.Errorf("failed to check fine status: %w", err)
	}

	if isPaid {
		return ErrFineAlreadyPaid
	}

	updateQuery := "UPDATE fines SET paid = TRUE WHERE id = $1"
	result, err := d.db.Exec(c, updateQuery, id)
	if err != nil {
		return fmt.Errorf("failed to pay fine: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrFineNotFound
	}

	return nil
}
