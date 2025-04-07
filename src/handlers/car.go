package handlers

import (
	"OPP/backend/api"
	"OPP/backend/dao"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type CarHandlers struct {
	dao dao.CarDao
}

func NewCarHandler() *CarHandlers {
	return &CarHandlers{
		dao: *dao.NewCarDao(),
	}
}

func (ch *CarHandlers) DeleteCars(c *gin.Context) {
	if err := ch.dao.DeleteAllCars(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete all cars"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "all cars deleted successfully"})
}

func (ch *CarHandlers) GetCars(c *gin.Context, params api.GetCarsParams) {
	cars := ch.dao.GetCars(c.Request.Context(), params.Limit, params.Offset, params.CurrentlyParked)
	c.JSON(http.StatusOK, cars)
}

func (ch *CarHandlers) DeleteUserCar(c *gin.Context, username string, plate string) {
	if err := ch.dao.DeleteUserCar(c.Request.Context(), username, plate); err != nil {
		if errors.Is(err, dao.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		if errors.Is(err, dao.ErrCarNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "car not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete car"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "car deleted successfully"})
}

func (ch *CarHandlers) GetUserCars(c *gin.Context, username string, params api.GetUserCarsParams) {
	cars := ch.dao.GetUserCars(c.Request.Context(), username, params.CurrentlyParked)
	c.JSON(http.StatusOK, cars)
}

func (ch *CarHandlers) UpdateUserCar(c *gin.Context, username string, plate string) {
	var car api.Car
	if err := c.ShouldBindJSON(&car); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if err := ch.dao.UpdateUserCar(c.Request.Context(), username, car); err != nil {
		if errors.Is(err, dao.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		if errors.Is(err, dao.ErrCarNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "car not found"})
			return
		}
		if errors.Is(err, dao.ErrCarAlreadyExists) {
			c.JSON(http.StatusConflict, gin.H{"error": "car already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update car"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "car updated successfully"})
}

func (ch *CarHandlers) AddUserCar(c *gin.Context, username string) {
	var car api.Car
	if err := c.ShouldBindJSON(&car); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if err := ch.dao.AddUserCar(c.Request.Context(), username, car); err != nil {
		if errors.Is(err, dao.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		if errors.Is(err, dao.ErrCarAlreadyExists) {
			c.JSON(http.StatusConflict, gin.H{"error": "car already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to add car"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "car added successfully"})
}
