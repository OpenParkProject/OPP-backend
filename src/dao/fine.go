package dao

import (
	"OPP/backend/api"
	"OPP/backend/db"
	"context"
	"errors"
	"fmt"
	"time"
)

var (
	ErrFineNotFound = errors.New("fine not found")
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
	query := "SELECT id, plate, amount, date FROM fines LIMIT $1 OFFSET $2"
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
		if err := rows.Scan(&fine.Id, &fine.Plate, &fine.Amount, &fine.Date); err != nil {
			fmt.Printf("row scan error: %v\n", err.Error())
			continue
		}
		fines = append(fines, fine)
	}

	return fines
}

func (d *FineDao) GetCarFines(c context.Context, plate string) []api.FineResponse {
	query := "SELECT id, plate, amount, date FROM fines WHERE plate = $1"
	rows, err := d.db.Query(c, query, plate)
	if err != nil {
		fmt.Printf("db error: %v\n", err.Error())
		return []api.FineResponse{}
	}
	defer rows.Close()

	fines := []api.FineResponse{}
	for rows.Next() {
		var fine api.FineResponse
		if err := rows.Scan(&fine.Id, &fine.Plate, &fine.Amount, &fine.Date); err != nil {
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

	query := "INSERT INTO fines (plate, amount, date) VALUES ($1, $2, $3) RETURNING id"
	currentDate := time.Now()
	var lastId int64
	err = d.db.QueryRow(c, query, plate, fine.Amount, currentDate).Scan(&lastId)
	if err != nil {
		return nil, fmt.Errorf("failed to add fine: %w", err)
	}

	return &api.FineResponse{
		Id:     lastId,
		Plate:  plate,
		Amount: fine.Amount,
		Date:   currentDate,
	}, nil
}

func (d *FineDao) DeleteFines(c context.Context) error {
	query := "DELETE FROM fines"
	_, err := d.db.Exec(c, query)
	if err != nil {
		return fmt.Errorf("failed to delete all fines: %w", err)
	}
	return nil
}

func (d *FineDao) GetUserFines(c context.Context, username string) ([]api.FineResponse, error) {
	query := "SELECT f.id, f.plate, f.amount, f.date FROM fines f JOIN cars c ON f.plate = c.plate WHERE c.user_id = $1"
	rows, err := d.db.Query(c, query, username)
	if err != nil {
		return nil, fmt.Errorf("failed to get user fines: %w", err)
	}
	defer rows.Close()

	fines := []api.FineResponse{}
	for rows.Next() {
		var fine api.FineResponse
		if err := rows.Scan(&fine.Id, &fine.Plate, &fine.Amount, &fine.Date); err != nil {
			return nil, fmt.Errorf("row scan error: %w", err)
		}
		fines = append(fines, fine)
	}

	return fines, nil
}
