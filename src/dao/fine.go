package dao

import (
	"OPP/backend/api"
	"OPP/backend/db"
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

func (d *FineDao) GetFines(limit *int, offset *int) []api.FineResponse {
	query := "SELECT id, plate, amount, date FROM fines LIMIT ? OFFSET ?"
	params := []any{20, 0}
	if limit != nil {
		params[0] = *limit
	}
	if offset != nil {
		params[1] = *offset
	}

	fines := []api.FineResponse{}
	rows, err := d.db.Query(query, params...)
	if err != nil {
		fmt.Printf("db error: %v\n", err.Error())
		return fines
	}
	defer rows.Close()

	for rows.Next() {
		var fine api.FineResponse
		var date string
		if err := rows.Scan(&fine.Id, &fine.Plate, &fine.Amount, &date); err != nil {
			fmt.Printf("row scan error: %v\n", err.Error())
			continue
		}

		parsedDate, err := time.Parse(time.RFC3339, date)
		if err != nil {
			fmt.Printf("failed to parse date: %v\n", err.Error())
			continue
		}
		fine.Date = parsedDate
		fines = append(fines, fine)
	}

	return fines
}

func (d *FineDao) GetCarFines(plate string) []api.FineResponse {
	query := "SELECT id, plate, amount, date FROM fines WHERE plate = ?"
	rows, err := d.db.Query(query, plate)
	if err != nil {
		fmt.Printf("db error: %v\n", err.Error())
		return []api.FineResponse{}
	}
	defer rows.Close()

	fines := []api.FineResponse{}
	for rows.Next() {
		var fine api.FineResponse
		var date string
		if err := rows.Scan(&fine.Id, &fine.Plate, &fine.Amount, &date); err != nil {
			fmt.Printf("row scan error: %v\n", err.Error())
			continue
		}

		parsedDate, err := time.Parse(time.RFC3339, date)
		if err != nil {
			fmt.Printf("failed to parse date: %v\n", err.Error())
			continue
		}
		fine.Date = parsedDate
		fines = append(fines, fine)
	}

	return fines
}

func (d *FineDao) AddCarFine(plate string, fine api.FineRequest) (*api.FineResponse, error) {
	carQuery := "SELECT * FROM cars WHERE plate = ?"
	carRows, err := d.db.Query(carQuery, plate)
	if err != nil {
		return nil, fmt.Errorf("failed to check car: %w", err)
	}
	defer carRows.Close()

	if !carRows.Next() {
		return nil, ErrCarNotFound
	}

	query := "INSERT INTO fines (plate, amount, date) VALUES (?, ?, ?)"
	currentDate := time.Now().Format(time.RFC3339)
	result, err := d.db.Exec(query, plate, fine.Amount, currentDate)
	if err != nil {
		return nil, fmt.Errorf("failed to add fine: %w", err)
	}

	lastId, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get fine ID: %w", err)
	}

	return &api.FineResponse{
		Id:     lastId,
		Plate:  plate,
		Amount: fine.Amount,
		Date:   time.Now(),
	}, nil
}

func (d *FineDao) DeleteFines() error {
	query := "DELETE FROM fines"
	_, err := d.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to delete all fines: %w", err)
	}
	return nil
}

func (d *FineDao) GetUserFines(username string) ([]api.FineResponse, error) {
	query := "SELECT f.id, f.plate, f.amount, f.date FROM fines f JOIN cars c ON f.plate = c.plate WHERE c.user_username = ?"
	rows, err := d.db.Query(query, username)
	if err != nil {
		return nil, fmt.Errorf("failed to get user fines: %w", err)
	}
	defer rows.Close()

	fines := []api.FineResponse{}
	for rows.Next() {
		var fine api.FineResponse
		var date string
		if err := rows.Scan(&fine.Id, &fine.Plate, &fine.Amount, &date); err != nil {
			return nil, fmt.Errorf("row scan error: %w", err)
		}

		parsedDate, err := time.Parse(time.RFC3339, date)
		if err != nil {
			return nil, fmt.Errorf("failed to parse date: %w", err)
		}
		fine.Date = parsedDate
		fines = append(fines, fine)
	}

	return fines, nil
}
