package dao

import (
	"OPP/backend/api"
	"OPP/backend/db"
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"
)

var (
	ErrTotemNotFound      = errors.New("totem not found")
	ErrTotemAlreadyExists = errors.New("totem already exists")
)

type TotemDao struct {
	db db.DB
}

func NewTotemDao() *TotemDao {
	return &TotemDao{
		db: *db.GetDB(),
	}
}

func (td *TotemDao) GetTotemById(ctx context.Context, id string) (error, api.TotemResponse) {
	query := `
        SELECT id, zone_id, latitude, longitude, registration_time 
        FROM totems 
        WHERE id = $1
    `

	var totem api.TotemResponse
	var registrationTime time.Time

	err := td.db.QueryRow(ctx, query, id).Scan(
		&totem.Id,
		&totem.ZoneId,
		&totem.Latitude,
		&totem.Longitude,
		&registrationTime,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrTotemNotFound, api.TotemResponse{}
		}
		return err, api.TotemResponse{}
	}

	totem.RegistrationTime = registrationTime

	return nil, totem
}

func (td *TotemDao) AddTotem(ctx context.Context, config api.TotemRequest) error {
	query := `
        INSERT INTO totems (id, zone_id, latitude, longitude) 
        VALUES ($1, $2, $3, $4)
        ON CONFLICT (id) DO UPDATE 
        SET zone_id = $2, latitude = $3, longitude = $4, registration_time = NOW()
    `

	_, err := td.db.Exec(ctx, query,
		config.Id,
		config.ZoneId,
		config.Latitude,
		config.Longitude,
	)

	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			return ErrTotemAlreadyExists
		}
	}
	return err
}

func (td *TotemDao) GetTotems(ctx context.Context, limit int, offset int) ([]api.TotemResponse, error) {
	query := `
				SELECT id, zone_id, latitude, longitude, registration_time 
				FROM totems
				LIMIT $1 OFFSET $2
		`

	rows, err := td.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var totems []api.TotemResponse

	for rows.Next() {
		var totem api.TotemResponse
		var registrationTime time.Time

		if err := rows.Scan(
			&totem.Id,
			&totem.ZoneId,
			&totem.Latitude,
			&totem.Longitude,
			&registrationTime,
		); err != nil {
			return nil, err
		}

		totem.RegistrationTime = registrationTime
		totems = append(totems, totem)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return totems, nil
}

func (td *TotemDao) DeleteTotemById(ctx context.Context, id string) error {
	query := `
				DELETE FROM totems 
				WHERE id = $1
		`

	result, err := td.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrTotemNotFound
	}

	return nil
}
