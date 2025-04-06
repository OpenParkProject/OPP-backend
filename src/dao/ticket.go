package dao

import (
	"OPP/backend/api"
	"OPP/backend/db"
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

func (d *TicketDao) GetTickets(limit *int, offset *int, validOnly *bool, startDateAfter *time.Time, endDateBefore *time.Time) []api.TicketResponse {
	query := "SELECT id, plate, start_date, end_date, price, paid, creation_time FROM tickets"
	var conditions []string
	var params []any

	if validOnly != nil && *validOnly {
		conditions = append(conditions, "paid = 1 AND end_date >= datetime('now')")
	}

	if startDateAfter != nil {
		conditions = append(conditions, "start_date >= ?")
		params = append(params, startDateAfter.Format(time.RFC3339))
	}

	if endDateBefore != nil {
		conditions = append(conditions, "end_date <= ?")
		params = append(params, endDateBefore.Format(time.RFC3339))
	}

	if len(conditions) > 0 {
		query += " WHERE " + conditions[0]
		for _, condition := range conditions[1:] {
			query += " AND " + condition
		}
	}

	query += " ORDER BY id DESC LIMIT ? OFFSET ?"
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
	rows, err := d.db.Query(query, params...)
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

func (d *TicketDao) GetTicketById(id int64) (*api.TicketResponse, error) {
	query := "SELECT id, plate, start_date, end_date, price, paid, creation_time FROM tickets WHERE id = ?"
	rows, err := d.db.Query(query, id)
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

func (d *TicketDao) AddCarTicket(plate string, ticket api.TicketRequest) (*api.TicketResponse, error) {
	carQuery := "SELECT * FROM cars WHERE plate = ?"
	carRows, err := d.db.Query(carQuery, plate)
	if err != nil {
		return nil, fmt.Errorf("failed to check car: %w", err)
	}
	defer carRows.Close()

	if !carRows.Next() {
		return nil, ErrCarNotFound
	}

	endTime := ticket.StartDate.Add(time.Duration(ticket.Duration) * time.Minute)
	price := float32(ticket.Duration) / 60.0

	query := "INSERT INTO tickets (plate, start_date, end_date, price, paid, creation_time) VALUES (?, ?, ?, ?, ?, datetime('now'))"
	result, err := d.db.Exec(query, plate, ticket.StartDate.Format(time.RFC3339), endTime.Format(time.RFC3339), price, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to add ticket: %w", err)
	}

	lastId, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get ticket ID: %w", err)
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

func (d *TicketDao) PayTicket(id int64) (*api.TicketResponse, error) {
	ticket, err := d.GetTicketById(id)
	if err != nil {
		return nil, err
	}

	if ticket.Paid == true {
		return nil, ErrTicketAlreadyPaid
	}

	query := "UPDATE tickets SET paid = 1 WHERE id = ?"
	_, err = d.db.Exec(query, id)
	if err != nil {
		return nil, fmt.Errorf("failed to update ticket: %w", err)
	}

	return d.GetTicketById(id)
}

func (d *TicketDao) GetCarTickets(plate string) ([]api.TicketResponse, error) {
	query := "SELECT id, plate, start_date, end_date, price, paid, creation_time FROM tickets WHERE plate = ?"
	rows, err := d.db.Query(query, plate)
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

func (d *TicketDao) GetUserTickets(username string, validOnly bool) ([]api.TicketResponse, error) {
	query := "SELECT t.id, t.plate, t.start_date, t.end_date, t.price, t.paid, t.creation_time FROM tickets AS t JOIN cars AS c ON t.plate = c.plate WHERE c.user_username = ?"
	if validOnly {
		query += " AND t.paid = 1 AND t.end_date >= datetime('now')"
	}

	rows, err := d.db.Query(query, username)
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
