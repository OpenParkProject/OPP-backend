package dao

import (
	"OPP/backend/api"
	"OPP/backend/db"
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
)

var (
	ErrZoneAlreadyExists         = errors.New("zone already exists")
	ErrZoneNotFound              = errors.New("zone not found")
	ErrZoneOverlap               = errors.New("zone overlaps with existing zone")
	ErrZoneUserRoleNotFound      = errors.New("zone user role not found")
	ErrZoneUserRoleAlreadyExists = errors.New("zone user role already exists")
	ErrZoneUserRoleInvalid       = errors.New("invalid zone user role")
)

type ZoneDao struct {
	db db.DB
}

func NewZoneDao() *ZoneDao {
	return &ZoneDao{
		db: *db.GetDB(),
	}
}

func (z *ZoneDao) CreateZone(c context.Context, zone api.ZoneRequest) (*api.ZoneResponse, error) {
	query := `
		INSERT INTO zones (
			name, 
			available, 
			geometry, 
			metadata, 
			price_offset, 
			price_lin, 
			price_exp
		) 
		VALUES ($1, $2, ST_GeomFromGeoJSON($3), $4, $5, $6, $7) 
		RETURNING 
			id, 
			name, 
			available, 
			ST_AsGeoJSON(geometry) as geometry, 
			metadata, 
			created_at, 
			updated_at, 
			price_offset, 
			price_lin, 
			price_exp
	`

	row := z.db.QueryRow(
		c,
		query,
		zone.Name,
		zone.Available,
		zone.Geometry, // GeoJSON string
		zone.Metadata,
		zone.PriceOffset,
		zone.PriceLin,
		zone.PriceExp,
	)

	var response api.ZoneResponse
	var geometryJSON string

	err := row.Scan(
		&response.Id,
		&response.Name,
		&response.Available,
		&geometryJSON,
		&response.Metadata,
		&response.CreatedAt,
		&response.UpdatedAt,
		&response.PriceOffset,
		&response.PriceLin,
		&response.PriceExp,
	)

	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return nil, ErrZoneAlreadyExists
		}
		if strings.Contains(err.Error(), "Zone overlaps with an existing zone") {
			return nil, ErrZoneOverlap
		}
		return nil, fmt.Errorf("failed to create zone: %w", err)
	}

	response.Geometry = geometryJSON
	return &response, nil
}

func (z *ZoneDao) GetAllZones(c context.Context) ([]api.ZoneResponse, error) {
	query := `
		SELECT 
			id, 
			name, 
			available, 
			ST_AsGeoJSON(geometry) as geometry, 
			metadata, 
			created_at, 
			updated_at, 
			price_offset, 
			price_lin, 
			price_exp
		FROM zones
	`

	rows, err := z.db.Query(c, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query zones: %w", err)
	}
	defer rows.Close()

	zones := []api.ZoneResponse{}
	for rows.Next() {
		var zone api.ZoneResponse
		var geometryJSON string

		if err := rows.Scan(
			&zone.Id,
			&zone.Name,
			&zone.Available,
			&geometryJSON,
			&zone.Metadata,
			&zone.CreatedAt,
			&zone.UpdatedAt,
			&zone.PriceOffset,
			&zone.PriceLin,
			&zone.PriceExp,
		); err != nil {
			return nil, fmt.Errorf("failed to scan zone: %w", err)
		}

		zone.Geometry = geometryJSON
		zones = append(zones, zone)
	}

	return zones, nil
}

func (z *ZoneDao) GetZoneById(c context.Context, id int64) (*api.ZoneResponse, error) {
	query := `
		SELECT 
			id, 
			name, 
			available, 
			ST_AsGeoJSON(geometry) as geometry, 
			metadata, 
			created_at, 
			updated_at, 
			price_offset, 
			price_lin, 
			price_exp
		FROM zones
		WHERE id = $1
	`

	row := z.db.QueryRow(c, query, id)

	var zone api.ZoneResponse
	var geometryJSON string

	if err := row.Scan(
		&zone.Id,
		&zone.Name,
		&zone.Available,
		&geometryJSON,
		&zone.Metadata,
		&zone.CreatedAt,
		&zone.UpdatedAt,
		&zone.PriceOffset,
		&zone.PriceLin,
		&zone.PriceExp,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrZoneNotFound
		}
		return nil, fmt.Errorf("failed to scan zone: %w", err)
	}

	zone.Geometry = geometryJSON
	return &zone, nil
}

func (z *ZoneDao) UpdateZone(c context.Context, id int64, zone api.ZoneRequest) (*api.ZoneResponse, error) {
	query := `
		UPDATE zones 
		SET 
			name = $1,
			available = $2,
			geometry = CASE WHEN $3::TEXT IS NOT NULL THEN ST_GeomFromGeoJSON($3) ELSE geometry END,
			metadata = $4,
			updated_at = NOW(),
			price_offset = $5,
			price_lin = $6,
			price_exp = $7
		WHERE id = $8
		RETURNING 
			id, 
			name, 
			available, 
			ST_AsGeoJSON(geometry) as geometry, 
			metadata, 
			created_at, 
			updated_at, 
			price_offset, 
			price_lin, 
			price_exp
	`

	row := z.db.QueryRow(
		c,
		query,
		zone.Name,
		zone.Available,
		zone.Geometry,
		zone.Metadata,
		zone.PriceOffset,
		zone.PriceLin,
		zone.PriceExp,
		id,
	)

	var updatedZone api.ZoneResponse
	var geometryJSON string

	if err := row.Scan(
		&updatedZone.Id,
		&updatedZone.Name,
		&updatedZone.Available,
		&geometryJSON,
		&updatedZone.Metadata,
		&updatedZone.CreatedAt,
		&updatedZone.UpdatedAt,
		&updatedZone.PriceOffset,
		&updatedZone.PriceLin,
		&updatedZone.PriceExp,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrZoneNotFound
		}

		// Check for overlap error
		if strings.Contains(err.Error(), "Zone overlaps with an existing zone") {
			return nil, ErrZoneOverlap
		}

		return nil, fmt.Errorf("failed to update zone: %w", err)
	}

	updatedZone.Geometry = geometryJSON
	return &updatedZone, nil
}

func (z *ZoneDao) DeleteZoneById(c context.Context, id int64) error {
	query := "DELETE FROM zones WHERE id = $1"
	result, err := z.db.Exec(c, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete zone: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrZoneNotFound
	}

	return nil
}

func (z *ZoneDao) GetZoneByName(c context.Context, name string) (*api.ZoneResponse, error) {
	query := `
		SELECT 
			id, 
			name, 
		available, 
			ST_AsGeoJSON(geometry) as geometry, 
			metadata, 
			created_at, 
			updated_at, 
			price_offset, 
			price_lin, 
			price_exp
		FROM zones
		WHERE name = $1
	`

	row := z.db.QueryRow(c, query, name)

	var zone api.ZoneResponse
	var geometryJSON string

	if err := row.Scan(
		&zone.Id,
		&zone.Name,
		&zone.Available,
		&geometryJSON,
		&zone.Metadata,
		&zone.CreatedAt,
		&zone.UpdatedAt,
		&zone.PriceOffset,
		&zone.PriceLin,
		&zone.PriceExp,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrZoneNotFound
		}
		return nil, fmt.Errorf("failed to scan zone: %w", err)
	}

	zone.Geometry = geometryJSON
	return &zone, nil
}

// Helper function to check if a zone exists
func (z *ZoneDao) ZoneExists(c context.Context, zoneId int64) (bool, error) {
	query := "SELECT 1 FROM zones WHERE id = $1"
	row := z.db.QueryRow(c, query, zoneId)

	var exists int
	if err := row.Scan(&exists); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

// FindAllZonesContainingPoint returns all zones that contain a given point
func (z *ZoneDao) FindAllZonesContainingPoint(c context.Context, longitude float64, latitude float64) ([]api.ZoneResponse, error) {
	query := `
        SELECT 
            id, 
            name, 
            available, 
            price_offset, 
            price_lin, 
            price_exp,
            ST_AsGeoJSON(geometry) AS geometry,
            metadata,
            created_at,
            updated_at
        FROM zones 
        WHERE ST_Contains(geometry, ST_SetSRID(ST_MakePoint($1, $2), 4326))
    `

	rows, err := z.db.Query(c, query, longitude, latitude)
	if err != nil {
		return nil, fmt.Errorf("failed to query zones: %w", err)
	}
	defer rows.Close()

	var zones []api.ZoneResponse
	for rows.Next() {
		var zone api.ZoneResponse
		var geometryJSON string

		if err := rows.Scan(
			&zone.Id,
			&zone.Name,
			&zone.Available,
			&zone.PriceOffset,
			&zone.PriceLin,
			&zone.PriceExp,
			&geometryJSON,
			&zone.Metadata,
			&zone.CreatedAt,
			&zone.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan zone: %w", err)
		}

		zone.Geometry = geometryJSON
		zones = append(zones, zone)
	}

	if len(zones) == 0 {
		return []api.ZoneResponse{}, nil
	}

	return zones, nil
}

// IsCoordinateInZone checks if a coordinate (point) is inside any zone
func (z *ZoneDao) IsCoordinateInZone(c context.Context, latitude float64, longitude float64) (*api.ZoneResponse, error) {
	query := `
        SELECT 
            id, 
            name, 
            available, 
            price_offset, 
            price_lin, 
            price_exp,
            ST_AsGeoJSON(geometry) AS geometry,
            metadata,
            created_at,
            updated_at
        FROM zones 
        WHERE ST_Contains(geometry, ST_SetSRID(ST_MakePoint($1, $2), 4326))
        LIMIT 1
    `

	row := z.db.QueryRow(c, query, longitude, latitude)

	var zone api.ZoneResponse
	var geometryJSON string

	if err := row.Scan(
		&zone.Id,
		&zone.Name,
		&zone.Available,
		&zone.PriceOffset,
		&zone.PriceLin,
		&zone.PriceExp,
		&geometryJSON,
		&zone.Metadata,
		&zone.CreatedAt,
		&zone.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrZoneNotFound
		}
		return nil, fmt.Errorf("failed to scan zone: %w", err)
	}

	zone.Geometry = geometryJSON
	return &zone, nil
}

// GetZoneUserRoles returns all user roles for a zone
func (z *ZoneDao) GetZoneUserRoles(c context.Context, zoneId int64) ([]api.ZoneUserRoleResponse, error) {
	query := `
        SELECT 
            id,
            zone_id,
            user_id,
            role,
            assigned_at,
            assigned_by
        FROM zone_user_roles
        WHERE zone_id = $1
    `

	rows, err := z.db.Query(c, query, zoneId)
	if err != nil {
		return nil, fmt.Errorf("failed to query zone user roles: %w", err)
	}
	defer rows.Close()

	var roles []api.ZoneUserRoleResponse
	for rows.Next() {
		var role api.ZoneUserRoleResponse

		if err := rows.Scan(
			&role.Id,
			&role.ZoneId,
			&role.Username,
			&role.Role,
			&role.AssignedAt,
			&role.AssignedBy,
		); err != nil {
			return nil, fmt.Errorf("failed to scan zone user role: %w", err)
		}

		roles = append(roles, role)
	}

	return roles, nil
}

// AddUserToZone adds a user role to a zone
func (z *ZoneDao) AddUserToZone(c context.Context, zoneId int64, request api.ZoneUserRoleRequest, assignedBy string) (*api.ZoneUserRoleResponse, error) {
	// First check if the zone exists
	zoneExists, err := z.ZoneExists(c, zoneId)
	if err != nil {
		return nil, fmt.Errorf("failed to check if zone exists: %w", err)
	}

	if !zoneExists {
		return nil, ErrZoneNotFound
	}

	query := `
        INSERT INTO zone_user_roles (
            zone_id,
            user_id,
            role,
            assigned_by
        )
        VALUES ($1, $2, $3, $4)
        RETURNING
            id,
            zone_id,
            user_id,
            role,
            assigned_at,
            assigned_by
    `

	row := z.db.QueryRow(c, query, zoneId, request.Username, request.Role, assignedBy)

	var userRole api.ZoneUserRoleResponse
	if err := row.Scan(
		&userRole.Id,
		&userRole.ZoneId,
		&userRole.Username,
		&userRole.Role,
		&userRole.AssignedAt,
		&userRole.AssignedBy,
	); err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return nil, ErrZoneUserRoleAlreadyExists
		}
		if strings.Contains(err.Error(), "invalid input value for enum zone_user_role") {
			return nil, ErrZoneUserRoleInvalid
		}
		return nil, fmt.Errorf("failed to add user to zone: %w", err)
	}

	return &userRole, nil
}

// RemoveUserFromZone removes a user role from a zone
func (z *ZoneDao) RemoveUserFromZone(c context.Context, zoneId int64, username string) error {
	query := "DELETE FROM zone_user_roles WHERE zone_id = $1 AND user_id = $2"
	result, err := z.db.Exec(c, query, zoneId, username)
	if err != nil {
		return fmt.Errorf("failed to remove user from zone: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrZoneNotFound
	}

	return nil
}

func (z *ZoneDao) CheckUserZonePermission(c context.Context, zoneId int64, username string) (string, error) {
	query := `
				SELECT role 
				FROM zone_user_roles 
				WHERE zone_id = $1 AND user_id = $2
		`

	row := z.db.QueryRow(c, query, zoneId, username)

	var role string
	if err := row.Scan(&role); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", ErrZoneUserRoleNotFound
		}
		return "", fmt.Errorf("failed to check user zone permission: %w", err)
	}

	return role, nil
}

func (z *ZoneDao) GetUserZones(c context.Context, username string) ([]api.ZoneResponse, error) {
	query := `
		SELECT 
			z.id, 
			z.name, 
			z.available, 
			ST_AsGeoJSON(z.geometry) as geometry, 
			z.metadata, 
			z.created_at, 
			z.updated_at, 
			z.price_offset, 
			z.price_lin, 
			z.price_exp
		FROM zones z
		JOIN zone_user_roles zur ON z.id = zur.zone_id
		WHERE zur.user_id = $1
	`

	rows, err := z.db.Query(c, query, username)
	if err != nil {
		return nil, fmt.Errorf("failed to query user zones: %w", err)
	}
	defer rows.Close()

	var zones []api.ZoneResponse
	for rows.Next() {
		var zone api.ZoneResponse
		var geometryJSON string

		if err := rows.Scan(
			&zone.Id,
			&zone.Name,
			&zone.Available,
			&geometryJSON,
			&zone.Metadata,
			&zone.CreatedAt,
			&zone.UpdatedAt,
			&zone.PriceOffset,
			&zone.PriceLin,
			&zone.PriceExp,
		); err != nil {
			return nil, fmt.Errorf("failed to scan user zone: %w", err)
		}

		zone.Geometry = geometryJSON
		zones = append(zones, zone)
	}

	return zones, nil
}
