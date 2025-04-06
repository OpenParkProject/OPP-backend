package dao

import (
	"OPP/backend/api"
	"OPP/backend/db"
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

func (d *CarDao) GetCars(limit *int, offset *int, currentlyParked *bool) []api.Car {
	query := "SELECT c.plate, c.brand, c.model FROM cars c"
	if currentlyParked != nil && *currentlyParked {
		query += " INNER JOIN tickets t ON c.plate = t.plate WHERE t.status = 'active'"
	}
	query += " LIMIT ? OFFSET ?"

	params := []any{20, 0}
	if limit != nil {
		params[0] = *limit
	}
	if offset != nil {
		params[1] = *offset
	}

	cars := []api.Car{}
	rows, err := d.db.Query(query, params...)
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

func (d *CarDao) GetUserCars(username string, currentlyParked *bool) []api.Car {
	query := "SELECT c.plate, c.brand, c.model FROM cars c WHERE c.user_username = ?"
	params := []any{username}
	if currentlyParked != nil && *currentlyParked {
		query = "SELECT c.plate, c.brand, c.model FROM cars c " +
			"INNER JOIN tickets t ON c.plate = t.plate " +
			"WHERE c.user_username = ? AND t.status = 'active'"
	}

	cars := []api.Car{}
	rows, err := d.db.Query(query, params...)
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

func (d *CarDao) AddUserCar(username string, car api.Car) error {
	userQuery := "SELECT * FROM users WHERE username = ?"
	userRows, err := d.db.Query(userQuery, username)
	if err != nil {
		return fmt.Errorf("failed to check user: %w", err)
	}
	defer userRows.Close()

	if !userRows.Next() {
		return ErrUserNotFound
	}

	query := "INSERT INTO cars (plate, brand, model, user_username) VALUES (?, ?, ?, ?)"
	_, err = d.db.Exec(query, car.Plate, car.Brand, car.Model, username)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return ErrCarAlreadyExists
		}
		return fmt.Errorf("failed to add car: %w", err)
	}

	return nil
}

func (d *CarDao) UpdateUserCar(username string, car api.Car) error {
	userQuery := "SELECT * FROM users WHERE username = ?"
	userRows, err := d.db.Query(userQuery, username)
	if err != nil {
		return fmt.Errorf("failed to check user: %w", err)
	}
	defer userRows.Close()

	if !userRows.Next() {
		return ErrUserNotFound
	}

	carQuery := "SELECT * FROM cars WHERE plate = ? AND user_username = ?"
	carRows, err := d.db.Query(carQuery, car.Plate, username)
	if err != nil {
		return fmt.Errorf("failed to check car: %w", err)
	}
	defer carRows.Close()

	if !carRows.Next() {
		return ErrCarNotFound
	}

	query := "UPDATE cars SET brand = ?, model = ? WHERE plate = ? AND user_username = ?"
	_, err = d.db.Exec(query, car.Brand, car.Model, car.Plate, username)
	if err != nil {
		return fmt.Errorf("failed to update car: %w", err)
	}

	return nil
}

func (d *CarDao) DeleteUserCar(username string, plate string) error {
	userQuery := "SELECT * FROM users WHERE username = ?"
	userRows, err := d.db.Query(userQuery, username)
	if err != nil {
		return fmt.Errorf("failed to check user: %w", err)
	}
	defer userRows.Close()

	if !userRows.Next() {
		return ErrUserNotFound
	}

	query := "DELETE FROM cars WHERE user_username = ? AND plate = ?"
	result, err := d.db.Exec(query, username, plate)
	if err != nil {
		return fmt.Errorf("failed to delete car: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return ErrCarNotFound
	}

	return nil
}

func (d *CarDao) DeleteAllCars() error {
	query := "DELETE FROM cars"
	_, err := d.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to delete all cars: %w", err)
	}

	return nil
}
