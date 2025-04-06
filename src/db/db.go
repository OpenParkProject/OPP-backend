package db

import (
	"database/sql"
	"fmt"
	"sync"

	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	db *sql.DB
	mu sync.Mutex
}

// Singleton instance of DB
var instance *DB

func Init() error {
	db, err := sql.Open("sqlite3", "db/db.sql")
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return fmt.Errorf("%w", err)
	}

	instance = &DB{
		db: db,
	}

	return nil
}

func GetDB() *DB {
	return instance
}

func (d *DB) Close() error {
	if d.db == nil {
		return nil
	} else {
		return d.db.Close()
	}
}

func (d *DB) Query(query string, args ...any) (*sql.Rows, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.db == nil {
		return nil, fmt.Errorf("db not initialized")
	} else {
		return d.db.Query(query, args...)
	}
}

func (d *DB) Exec(query string, args ...any) (sql.Result, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.db == nil {
		return nil, fmt.Errorf("db not initialized")
	} else {
		return d.db.Exec(query, args...)
	}
}
