package handlers

import (
	"OPP/backend/api"
	"OPP/backend/auth"
	"OPP/backend/dao"
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type TicketHandlers struct {
	dao dao.TicketDao
}

func NewTicketHandler() *TicketHandlers {
	return &TicketHandlers{
		dao: *dao.NewTicketDao(),
	}
}

func ValidateTicketRequest(c context.Context, req api.TicketRequest) error {
	now := time.Now()

	if req.StartDate.Before(now) {
		return fmt.Errorf("start_date must be in the future")
	}

	endDate := req.StartDate.Add(time.Duration(req.Duration) * time.Minute)
	if endDate.Before(now) {
		return fmt.Errorf("end_date must be in the future")
	}

	if req.Duration <= 0 {
		return fmt.Errorf("duration must be greater than zero")
	}

	res, err := dao.NewZoneDao().ZoneExists(c, req.ZoneId)
	if err != nil {
		return fmt.Errorf("failed to check zone existence: %w", err)
	}
	if !res {
		return fmt.Errorf("zone with id %d does not exist", req.ZoneId)
	}

	return nil
}

func (th *TicketHandlers) GetTickets(c *gin.Context, params api.GetTicketsParams) {
	_, role, err := auth.GetPermissions(c)
	if err != nil {
		return
	}
	if role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	tickets := th.dao.GetTickets(c.Request.Context(), params.Limit, params.Offset, params.ValidOnly, params.StartDateAfter, params.EndDateBefore)
	c.JSON(http.StatusOK, tickets)
}

func (th *TicketHandlers) GetTicketById(c *gin.Context, id int64) {
	_, role, err := auth.GetPermissions(c)
	if err != nil {
		return
	}
	if role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	ticket, err := th.dao.GetTicketById(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, dao.ErrTicketNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "ticket not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get ticket"})
		return
	}

	c.JSON(http.StatusOK, ticket)
}

func (th *TicketHandlers) AddCarTicket(c *gin.Context, plate string) {
	var ticketRequest api.TicketRequest
	if err := c.ShouldBindJSON(&ticketRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	err := ValidateTicketRequest(c, ticketRequest)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ticket, err := th.dao.AddCarTicket(c.Request.Context(), plate, ticketRequest)
	if err != nil {
		if errors.Is(err, dao.ErrCarNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "car not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to add ticket"})
		return
	}

	c.JSON(http.StatusCreated, ticket)
}

func (th *TicketHandlers) GetCarTickets(c *gin.Context, plate string) {
	tickets, err := th.dao.GetCarTickets(c.Request.Context(), plate)
	if err != nil {
		if errors.Is(err, dao.ErrCarNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "car not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get tickets"})
		return
	}
	c.JSON(http.StatusOK, tickets)
}

func (th *TicketHandlers) PayTicket(c *gin.Context, id int64) {
	ticket, err := th.dao.PayTicket(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, dao.ErrTicketAlreadyPaid) {
			c.JSON(http.StatusConflict, gin.H{"error": "ticket already paid"})
			return
		}
		if errors.Is(err, dao.ErrTicketNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "ticket not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to pay ticket"})
		return
	}

	c.JSON(http.StatusOK, ticket)
}

func (th *TicketHandlers) GetUserTickets(c *gin.Context, params api.GetUserTicketsParams) {
	username, _, err := auth.GetPermissions(c)
	if err != nil {
		return
	}

	tickets, err := th.dao.GetUserTickets(c.Request.Context(), username, *params.ValidOnly)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user tickets"})
		return
	}

	c.JSON(http.StatusOK, tickets)

}

func (th *TicketHandlers) DeleteTicketById(c *gin.Context, id int64) {
	username, _, err := auth.GetPermissions(c)
	if err != nil {
		return
	}

	err = th.dao.DeleteTicketById(c.Request.Context(), username, id)
	if err != nil {
		if errors.Is(err, dao.ErrTicketNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "ticket not found"})
			return
		}
		if errors.Is(err, dao.ErrTicketNotOwned) {
			c.JSON(http.StatusForbidden, gin.H{"error": "ticket not owned by user"})
			return
		}
		if errors.Is(err, dao.ErrTicketAlreadyPaid) {
			c.JSON(http.StatusConflict, gin.H{"error": "ticket already paid"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete ticket"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "ticket deleted successfully"})
}
