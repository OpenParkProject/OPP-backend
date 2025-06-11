package dao

import (
	"OPP/backend/api"
	"OPP/backend/db"
	"context"
	"errors"
	"fmt"
	"math"
	"time"
)

var (
	ErrTicketNotFound    = errors.New("ticket not found")
	ErrTicketAlreadyPaid = errors.New("ticket already paid")
	ErrTicketNotOwned    = errors.New("ticket not owned by user")
)

type TicketDao struct {
	db db.DB
}

func NewTicketDao() *TicketDao {
	return &TicketDao{
		db: *db.GetDB(),
	}
}

func (d *TicketDao) GetTickets(c context.Context, limit *int, offset *int, validOnly *bool, startDateAfter *time.Time, endDateBefore *time.Time) []api.TicketResponse {
	query := "SELECT id, plate, start_date, end_date, price, paid, creation_time, zone_id FROM tickets"
	var conditions []string
	var params []any

	if validOnly != nil && *validOnly {
		conditions = append(conditions, "paid = TRUE AND end_date >= NOW()")
	}

	if startDateAfter != nil {
		conditions = append(conditions, "start_date >= $1")
		params = append(params, startDateAfter)
	}

	if endDateBefore != nil {
		conditions = append(conditions, "end_date <= $2")
		params = append(params, endDateBefore)
	}

	if len(conditions) > 0 {
		query += " WHERE " + conditions[0]
		for _, condition := range conditions[1:] {
			query += " AND " + condition
		}
	}

	query += " ORDER BY id DESC LIMIT $1 OFFSET $2"
	limitVal := 20
	offsetVal := 0
	if limit != nil {
		limitVal = *limit
	}
	if offset != nil {
		offsetVal = *offset
	}
	params = append(params, limitVal, offsetVal)

	tickets := []api.TicketResponse{}
	rows, err := d.db.Query(c, query, params...)
	if err != nil {
		return tickets
	}
	defer rows.Close()

	// Update the scan to include zone_id
	for rows.Next() {
		var ticket api.TicketResponse
		if err := rows.Scan(&ticket.Id, &ticket.Plate, &ticket.StartDate, &ticket.EndDate, &ticket.Price, &ticket.Paid, &ticket.CreationTime, &ticket.ZoneId); err != nil {
			continue
		}
		tickets = append(tickets, ticket)
	}

	return tickets
}

func (d *TicketDao) GetTicketById(c context.Context, id int64) (*api.TicketResponse, error) {
	query := "SELECT id, plate, start_date, end_date, price, paid, creation_time, zone_id FROM tickets WHERE id = $1"
	rows, err := d.db.Query(c, query, id)
	if err != nil {
		return nil, fmt.Errorf("failed to query ticket: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, ErrTicketNotFound
	}

	var ticket api.TicketResponse
	if err := rows.Scan(&ticket.Id, &ticket.Plate, &ticket.StartDate, &ticket.EndDate, &ticket.Price, &ticket.Paid, &ticket.CreationTime, &ticket.ZoneId); err != nil {
		return nil, fmt.Errorf("failed to scan ticket: %w", err)
	}

	return &ticket, nil
}

func (d *TicketDao) CreateZoneTicket(c context.Context, zoneId int64, ticket api.TicketRequest) (*api.TicketResponse, error) {
	carQuery := "SELECT * FROM cars WHERE plate = $1"
	carRows, err := d.db.Query(c, carQuery, ticket.Plate)
	if err != nil {
		return nil, fmt.Errorf("failed to check car: %w", err)
	}
	defer carRows.Close()

	if !carRows.Next() {
		return nil, ErrCarNotFound
	}

	endTime := ticket.StartDate.Add(time.Duration(ticket.Duration) * time.Minute)

	// Price calculation based on zone
	zone, err := NewZoneDao().GetZoneById(c, zoneId)
	if err != nil {
		return nil, fmt.Errorf("failed to get zone: %w", err)
	}

	durationInHours := float32(ticket.Duration) / 60.0
	priceComponent := zone.PriceLin * durationInHours
	price := zone.PriceOffset + float32(math.Pow(float64(priceComponent), float64(zone.PriceExp)))

	creationTime := time.Now()
	query := "INSERT INTO tickets (plate, start_date, end_date, price, paid, creation_time, zone_id) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id"
	var lastId int64
	err = d.db.QueryRow(c, query, ticket.Plate, ticket.StartDate, endTime, price, false, creationTime, zoneId).Scan(&lastId)
	if err != nil {
		return nil, fmt.Errorf("failed to add ticket: %w", err)
	}

	return &api.TicketResponse{
		Id:           lastId,
		Plate:        ticket.Plate,
		StartDate:    ticket.StartDate,
		EndDate:      endTime,
		Price:        price,
		Paid:         false,
		CreationTime: creationTime,
		ZoneId:       zoneId,
	}, nil
}

func (d *TicketDao) PayTicket(c context.Context, id int64) (*api.TicketResponse, error) {
	ticket, err := d.GetTicketById(c, id)
	if err != nil {
		return nil, err
	}

	if ticket.Paid == true {
		return nil, ErrTicketAlreadyPaid
	}

	query := "UPDATE tickets SET paid = TRUE WHERE id = $1"
	_, err = d.db.Exec(c, query, id)
	if err != nil {
		return nil, fmt.Errorf("failed to update ticket: %w", err)
	}

	return d.GetTicketById(c, id)
}

func (d *TicketDao) GetCarTickets(c context.Context, plate string) ([]api.TicketResponse, error) {
	query := "SELECT id, plate, start_date, end_date, price, paid, creation_time, zone_id FROM tickets WHERE plate = $1"
	rows, err := d.db.Query(c, query, plate)
	if err != nil {
		return nil, fmt.Errorf("failed to query tickets: %w", err)
	}
	defer rows.Close()

	tickets := []api.TicketResponse{}
	for rows.Next() {
		var ticket api.TicketResponse
		if err := rows.Scan(&ticket.Id, &ticket.Plate, &ticket.StartDate, &ticket.EndDate, &ticket.Price, &ticket.Paid, &ticket.CreationTime, &ticket.ZoneId); err != nil {
			return nil, fmt.Errorf("failed to scan ticket: %w", err)
		}
		tickets = append(tickets, ticket)
	}

	return tickets, nil
}

func (d *TicketDao) GetUserTickets(c context.Context, username string, validOnly bool) ([]api.TicketResponse, error) {
	query := "SELECT t.id, t.plate, t.start_date, t.end_date, t.price, t.paid, t.creation_time, t.zone_id FROM tickets AS t JOIN cars AS c ON t.plate = c.plate WHERE c.user_id = $1"
	if validOnly {
		query += " AND t.paid = TRUE AND t.end_date >= NOW()"
	}

	rows, err := d.db.Query(c, query, username)
	if err != nil {
		return nil, fmt.Errorf("failed to query user tickets: %w", err)
	}
	defer rows.Close()

	tickets := []api.TicketResponse{}
	for rows.Next() {
		var ticket api.TicketResponse
		if err := rows.Scan(&ticket.Id, &ticket.Plate, &ticket.StartDate, &ticket.EndDate, &ticket.Price, &ticket.Paid, &ticket.CreationTime, &ticket.ZoneId); err != nil {
			return nil, fmt.Errorf("failed to scan ticket: %w", err)
		}
		tickets = append(tickets, ticket)
	}

	return tickets, nil
}

func (d *TicketDao) DeleteTicketById(c context.Context, username string, id int64) error {
	ticket, err := d.GetTicketById(c, id)
	if err != nil {
		if errors.Is(err, ErrTicketNotFound) {
			return ErrTicketNotFound
		}
		return fmt.Errorf("failed to get ticket: %w", err)
	}

	if ticket.Paid {
		return ErrTicketAlreadyPaid
	}

	// Check if the user owns the ticket
	query := "SELECT 1 FROM cars WHERE plate = $1 AND user_id = $2"
	rows, err := d.db.Query(c, query, ticket.Plate, username)
	if err != nil {
		return fmt.Errorf("failed to check ticket ownership: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return ErrTicketNotOwned
	}

	deleteQuery := "DELETE FROM tickets WHERE id = $1"
	result, err := d.db.Exec(c, deleteQuery, id)
	if err != nil {
		return fmt.Errorf("failed to delete ticket: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrTicketNotFound
	}

	return nil
}

func (d *FineDao) GetZoneTickets(ctx context.Context, zoneId int64, limit int, offset int) []api.TicketResponse {
	query := "SELECT id, plate, start_date, end_date, price, paid, creation_time, zone_id FROM tickets WHERE zone_id = $1 ORDER BY id DESC LIMIT $2 OFFSET $3"
	rows, err := d.db.Query(ctx, query, zoneId, limit, offset)
	if err != nil {
		return nil
	}
	defer rows.Close()

	tickets := []api.TicketResponse{}
	for rows.Next() {
		var ticket api.TicketResponse
		if err := rows.Scan(&ticket.Id, &ticket.Plate, &ticket.StartDate, &ticket.EndDate, &ticket.Price, &ticket.Paid, &ticket.CreationTime, &ticket.ZoneId); err != nil {
			continue
		}
		tickets = append(tickets, ticket)
	}

	return tickets
}
