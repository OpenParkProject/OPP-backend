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
	// Auth middleware sets values in request context
	// not in gin context
	username := c.Request.Context().Value("username")
	if username == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	_, ok := username.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get username"})
		return
	}
	role := c.Request.Context().Value("role")
	if role == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	roleStr, ok := role.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get role"})
		return
	}
	if roleStr != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	if err := ch.dao.DeleteAllCars(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete all cars"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "all cars deleted successfully"})
}

func (ch *CarHandlers) GetCars(c *gin.Context, params api.GetCarsParams) {
	username := c.Request.Context().Value("username")
	if username == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	_, ok := username.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get username"})
		return
	}
	role := c.Request.Context().Value("role")
	if role == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	roleStr, ok := role.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get role"})
		return
	}
	if roleStr != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	cars := ch.dao.GetCars(c.Request.Context(), params.Limit, params.Offset, params.CurrentlyParked)
	c.JSON(http.StatusOK, cars)
}

func (ch *CarHandlers) DeleteUserCar(c *gin.Context, plate string) {
	username := c.Request.Context().Value("username")
	if username == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	usernamestr, ok := username.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get username"})
		return
	}
	role := c.Request.Context().Value("role")
	if role == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	_, ok = role.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get role"})
		return
	}

	if err := ch.dao.DeleteUserCar(c.Request.Context(), usernamestr, plate); err != nil {
		if errors.Is(err, dao.ErrCarNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "car not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete car"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "car deleted successfully"})
}

func (ch *CarHandlers) GetUserCars(c *gin.Context, params api.GetUserCarsParams) {
	username := c.Request.Context().Value("username")
	if username == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	usernamestr, ok := username.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get username"})
		return
	}
	role := c.Request.Context().Value("role")
	if role == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	_, ok = role.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get role"})
		return
	}

	cars := ch.dao.GetUserCars(c.Request.Context(), usernamestr, params.CurrentlyParked)
	c.JSON(http.StatusOK, cars)
}

func (ch *CarHandlers) UpdateUserCar(c *gin.Context, plate string) {
	username := c.Request.Context().Value("username")
	if username == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	usernamestr, ok := username.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get username"})
		return
	}
	role := c.Request.Context().Value("role")
	if role == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	_, ok = role.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get role"})
		return
	}

	var car api.Car
	if err := c.ShouldBindJSON(&car); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if err := ch.dao.UpdateUserCar(c.Request.Context(), usernamestr, car); err != nil {
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

func (ch *CarHandlers) AddUserCar(c *gin.Context) {
	username := c.Request.Context().Value("username")
	if username == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	usernamestr, ok := username.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get username"})
		return
	}
	role := c.Request.Context().Value("role")
	if role == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	_, ok = role.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get role"})
		return
	}

	var car api.Car
	if err := c.ShouldBindJSON(&car); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if err := ch.dao.AddUserCar(c.Request.Context(), usernamestr, car); err != nil {
		if errors.Is(err, dao.ErrCarAlreadyExists) {
			c.JSON(http.StatusConflict, gin.H{"error": "car already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to add car"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "car added successfully"})
}
