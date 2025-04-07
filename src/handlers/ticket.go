package handlers

import (
	"OPP/backend/api"
	"OPP/backend/dao"
	"errors"
	"net/http"

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

func (th *TicketHandlers) GetTickets(c *gin.Context, params api.GetTicketsParams) {
	tickets := th.dao.GetTickets(c.Request.Context(), params.Limit, params.Offset, params.ValidOnly, params.StartDateAfter, params.EndDateBefore)
	c.JSON(http.StatusOK, tickets)
}

func (th *TicketHandlers) GetTicketById(c *gin.Context, id int64) {
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

func (th *TicketHandlers) GetUserTickets(c *gin.Context, username string, params api.GetUserTicketsParams) {
	tickets, err := th.dao.GetUserTickets(c.Request.Context(), username, *params.ValidOnly)
	if err != nil {
		if errors.Is(err, dao.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user tickets"})
		return
	}

	c.JSON(http.StatusOK, tickets)

}
