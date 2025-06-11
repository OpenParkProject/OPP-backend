package dao

import (
	"OPP/backend/api"
	"OPP/backend/db"
	"context"
	"database/sql"
	"errors"
	"time"
)

var (
	ErrTotemNotFound = errors.New("totem not found")
)

type TotemDao struct {
	db db.DB
}

func NewTotemDao() *TotemDao {
	return &TotemDao{
		db: *db.GetDB(),
	}
}

func (td *TotemDao) GetTotemById(ctx context.Context, id int64) (error, api.TotemResponse) {
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

	return err
}
