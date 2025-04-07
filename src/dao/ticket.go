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
	ErrTicketNotFound    = errors.New("ticket not found")
	ErrTicketAlreadyPaid = errors.New("ticket already paid")
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
	query := "SELECT id, plate, start_date, end_date, price, paid, creation_time FROM tickets"
	var conditions []string
	var params []any

	if validOnly != nil && *validOnly {
		conditions = append(conditions, "paid = 1 AND end_date >= datetime('now')")
	}

	if startDateAfter != nil {
		conditions = append(conditions, "start_date >= $1")
		params = append(params, startDateAfter.Format(time.RFC3339))
	}

	if endDateBefore != nil {
		conditions = append(conditions, "end_date <= $2")
		params = append(params, endDateBefore.Format(time.RFC3339))
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
		fmt.Printf("db error %v\n", err.Error())
		return tickets
	}
	defer rows.Close()

	for rows.Next() {
		var ticket api.TicketResponse
		if err := rows.Scan(&ticket.Id, &ticket.Plate, &ticket.StartDate, &ticket.EndDate, &ticket.Price, &ticket.Paid, &ticket.CreationTime); err != nil {
			fmt.Printf("row scan error %v\n", err.Error())
			continue
		}
		tickets = append(tickets, ticket)
	}

	return tickets
}

func (d *TicketDao) GetTicketById(c context.Context, id int64) (*api.TicketResponse, error) {
	query := "SELECT id, plate, start_date, end_date, price, paid, creation_time FROM tickets WHERE id = $1"
	rows, err := d.db.Query(c, query, id)
	if err != nil {
		return nil, fmt.Errorf("failed to query ticket: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, ErrTicketNotFound
	}

	var ticket api.TicketResponse
	if err := rows.Scan(&ticket.Id, &ticket.Plate, &ticket.StartDate, &ticket.EndDate, &ticket.Price, &ticket.Paid, &ticket.CreationTime); err != nil {
		return nil, fmt.Errorf("failed to scan ticket: %w", err)
	}

	return &ticket, nil
}

func (d *TicketDao) AddCarTicket(c context.Context, plate string, ticket api.TicketRequest) (*api.TicketResponse, error) {
	carQuery := "SELECT * FROM cars WHERE plate = $1"
	carRows, err := d.db.Query(c, carQuery, plate)
	if err != nil {
		return nil, fmt.Errorf("failed to check car: %w", err)
	}
	defer carRows.Close()

	if !carRows.Next() {
		return nil, ErrCarNotFound
	}

	endTime := ticket.StartDate.Add(time.Duration(ticket.Duration) * time.Minute)
	price := float32(ticket.Duration) / 60.0

	query := "INSERT INTO tickets (plate, start_date, end_date, price, paid, creation_time) VALUES ($1, $2, $3, $4, $5, NOW()) RETURNING id"
	var lastId int64
	err = d.db.QueryRow(c, query, plate, ticket.StartDate.Format(time.RFC3339), endTime.Format(time.RFC3339), price, 0).Scan(&lastId)
	if err != nil {
		return nil, fmt.Errorf("failed to add ticket: %w", err)
	}

	return &api.TicketResponse{
		Id:           lastId,
		Plate:        plate,
		StartDate:    ticket.StartDate,
		EndDate:      endTime,
		Price:        price,
		Paid:         false,
		CreationTime: time.Now(),
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

	query := "UPDATE tickets SET paid = 1 WHERE id = $1"
	_, err = d.db.Exec(c, query, id)
	if err != nil {
		return nil, fmt.Errorf("failed to update ticket: %w", err)
	}

	return d.GetTicketById(c, id)
}

func (d *TicketDao) GetCarTickets(c context.Context, plate string) ([]api.TicketResponse, error) {
	query := "SELECT id, plate, start_date, end_date, price, paid, creation_time FROM tickets WHERE plate = $1"
	rows, err := d.db.Query(c, query, plate)
	if err != nil {
		return nil, fmt.Errorf("failed to query tickets: %w", err)
	}
	defer rows.Close()

	tickets := []api.TicketResponse{}
	for rows.Next() {
		var ticket api.TicketResponse
		if err := rows.Scan(&ticket.Id, &ticket.Plate, &ticket.StartDate, &ticket.EndDate, &ticket.Price, &ticket.Paid, &ticket.CreationTime); err != nil {
			return nil, fmt.Errorf("failed to scan ticket: %w", err)
		}
		tickets = append(tickets, ticket)
	}

	return tickets, nil
}

func (d *TicketDao) GetUserTickets(c context.Context, username string, validOnly bool) ([]api.TicketResponse, error) {
	query := "SELECT t.id, t.plate, t.start_date, t.end_date, t.price, t.paid, t.creation_time FROM tickets AS t JOIN cars AS c ON t.plate = c.plate WHERE c.user_username = $1"
	if validOnly {
		query += " AND t.paid = 1 AND t.end_date >= datetime('now')"
	}

	rows, err := d.db.Query(c, query, username)
	if err != nil {
		return nil, fmt.Errorf("failed to query user tickets: %w", err)
	}
	defer rows.Close()

	tickets := []api.TicketResponse{}
	for rows.Next() {
		var ticket api.TicketResponse
		if err := rows.Scan(&ticket.Id, &ticket.Plate, &ticket.StartDate, &ticket.EndDate, &ticket.Price, &ticket.Paid, &ticket.CreationTime); err != nil {
			return nil, fmt.Errorf("failed to scan ticket: %w", err)
		}
		tickets = append(tickets, ticket)
	}

	return tickets, nil
}
