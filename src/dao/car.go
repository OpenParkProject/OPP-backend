package dao

import (
	"OPP/backend/api"
	"OPP/backend/db"
	"context"
	"errors"
	"fmt"
	"strings"
)

var (
	ErrCarAlreadyExists = errors.New("car already exists")
	ErrCarNotFound      = errors.New("car not found")
)

type CarDao struct {
	db db.DB
}

func NewCarDao() *CarDao {
	return &CarDao{
		db: *db.GetDB(),
	}
}

func (d *CarDao) GetCars(c context.Context, limit *int, offset *int, currentlyParked *bool) []api.Car {
	query := "SELECT c.plate, c.brand, c.model FROM cars c"
	if currentlyParked != nil && *currentlyParked {
		query += " INNER JOIN tickets t ON c.plate = t.plate WHERE t.status = 'active'"
	}
	query += " LIMIT $1 OFFSET $2"

	params := []any{20, 0}
	if limit != nil {
		params[0] = *limit
	}
	if offset != nil {
		params[1] = *offset
	}

	cars := []api.Car{}
	rows, err := d.db.Query(c, query, params...)
	if err != nil {
		fmt.Printf("db error: %v\n", err.Error())
		return cars
	}
	defer rows.Close()

	for rows.Next() {
		var car api.Car
		if err := rows.Scan(&car.Plate, &car.Brand, &car.Model); err != nil {
			fmt.Printf("row scan error: %v\n", err.Error())
			continue
		}
		cars = append(cars, car)
	}

	return cars
}

func (d *CarDao) GetUserCars(c context.Context, username string, currentlyParked *bool) []api.Car {
	query := "SELECT c.plate, c.brand, c.model FROM cars c WHERE c.user_id = $1"
	params := []any{username}
	if currentlyParked != nil && *currentlyParked {
		query = "SELECT c.plate, c.brand, c.model FROM cars c " +
			"INNER JOIN tickets t ON c.plate = t.plate " +
			"WHERE c.user_id = ? AND t.status = 'active'"
	}

	cars := []api.Car{}
	rows, err := d.db.Query(c, query, params...)
	if err != nil {
		fmt.Printf("db error: %v\n", err.Error())
		return cars
	}
	defer rows.Close()

	for rows.Next() {
		var car api.Car
		if err := rows.Scan(&car.Plate, &car.Brand, &car.Model); err != nil {
			fmt.Printf("row scan error: %v\n", err.Error())
			continue
		}
		cars = append(cars, car)
	}

	return cars
}

func (d *CarDao) AddUserCar(c context.Context, username string, car api.Car) error {
	query := "INSERT INTO cars (plate, brand, model, user_id) VALUES ($1, $2, $3, $4)"
	_, err := d.db.Exec(c, query, car.Plate, car.Brand, car.Model, username)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return ErrCarAlreadyExists
		}
		return fmt.Errorf("failed to add car: %w", err)
	}

	return nil
}

func (d *CarDao) UpdateUserCar(c context.Context, username string, car api.Car) error {
	carQuery := "SELECT * FROM cars WHERE plate = $1 AND user_id = $2"
	carRows, err := d.db.Query(c, carQuery, car.Plate, username)
	if err != nil {
		return fmt.Errorf("failed to check car: %w", err)
	}
	defer carRows.Close()

	if !carRows.Next() {
		return ErrCarNotFound
	}

	query := "UPDATE cars SET brand = $1, model = $2 WHERE plate = $3 AND user_id = $4"
	_, err = d.db.Exec(c, query, car.Brand, car.Model, car.Plate, username)
	if err != nil {
		return fmt.Errorf("failed to update car: %w", err)
	}

	return nil
}

func (d *CarDao) DeleteUserCar(c context.Context, username string, plate string) error {
	query := "DELETE FROM cars WHERE user_id = $1 AND plate = $2"
	result, err := d.db.Exec(c, query, username, plate)
	if err != nil {
		return fmt.Errorf("failed to delete car: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrCarNotFound
	}

	return nil
}

func (d *CarDao) DeleteAllCars(c context.Context) error {
	query := "DELETE FROM cars"
	_, err := d.db.Exec(c, query)
	if err != nil {
		return fmt.Errorf("failed to delete all cars: %w", err)
	}

	return nil
}
