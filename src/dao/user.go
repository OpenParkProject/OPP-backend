package dao

import (
	"OPP/backend/api"
	"OPP/backend/db"
	"errors"
	"fmt"
	"strings"
)

var (
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrInvalidUser       = errors.New("invalid user data")
	ErrUserNotFound      = errors.New("user not found")
)

type UserDao struct {
	db db.DB
}

func NewUserDao() *UserDao {
	return &UserDao{
		db: *db.GetDB(),
	}
}

func (d *UserDao) GetUsers(limit *int, offset *int) []api.UserResponse {
	query := "SELECT username, name, surname, email FROM users LIMIT ? OFFSET ?"
	params := []any{20, 0}
	if limit != nil {
		params[0] = *limit
	}
	if offset != nil {
		params[1] = *offset
	}

	var users []api.UserResponse
	rows, err := d.db.Query(query, params...)
	if err != nil {
		fmt.Printf("db error: %v\n", err.Error())
		return users
	}
	defer rows.Close()

	for rows.Next() {
		var user api.UserResponse
		if err := rows.Scan(&user.Username, &user.Name, &user.Surname, &user.Email); err != nil {
			fmt.Printf("row scan error: %v\n", err.Error())
			continue
		}
		users = append(users, user)
	}
	return users
}

func (d *UserDao) AddUser(user api.UserRequest) error {
	query := "INSERT INTO users (username, name, surname, email, password, role) VALUES (?, ?, ?, ?, ?, ?)"
	_, err := d.db.Exec(query, user.Username, user.Name, user.Surname, user.Email, user.Password, user.Role)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return ErrUserAlreadyExists
		}
		return fmt.Errorf("failed to add user: %w", err)
	}
	return nil
}

func (d *UserDao) GetUser(username string) (*api.UserResponse, error) {
	query := "SELECT username, name, surname, email FROM users WHERE username = ?"
	rows, err := d.db.Query(query, username)
	if err != nil {
		return nil, fmt.Errorf("db error: %w", err)
	}
	defer rows.Close()

	var user api.UserResponse
	if rows.Next() {
		if err := rows.Scan(&user.Username, &user.Name, &user.Surname, &user.Email); err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		return &user, nil
	}
	return nil, ErrUserNotFound
}

func (d *UserDao) DeleteAllUsers() error {
	query := "DELETE FROM users"
	_, err := d.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to delete all users: %w", err)
	}
	return nil
}

func (d *UserDao) DeleteUser(username string) error {
	query := "DELETE FROM users WHERE username = ?"
	result, err := d.db.Exec(query, username)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return ErrUserNotFound
	}
	return nil
}

func (d *UserDao) UpdateUser(username string, user api.UserRequest) error {
	query := "UPDATE users SET name = ?, surname = ?, email = ? WHERE username = ?"
	result, err := d.db.Exec(query, user.Name, user.Surname, user.Email, username)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return ErrUserAlreadyExists
		}
		return fmt.Errorf("failed to update user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return ErrUserNotFound
	}
	return nil
}
