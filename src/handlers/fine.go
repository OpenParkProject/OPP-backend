package handlers

import (
	"OPP/backend/api"
	"OPP/backend/dao"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type FineHandlers struct {
	dao dao.FineDao
}

func NewFineHandler() *FineHandlers {
	return &FineHandlers{
		dao: *dao.NewFineDao(),
	}
}

func (fh *FineHandlers) GetFines(c *gin.Context, params api.GetFinesParams) {
	fines := fh.dao.GetFines(c.Request.Context(), params.Limit, params.Offset)
	c.JSON(http.StatusOK, fines)
}

func (fh *FineHandlers) GetCarFines(c *gin.Context, plate string) {
	fines := fh.dao.GetCarFines(c.Request.Context(), plate)
	c.JSON(http.StatusOK, fines)
}

func (fh *FineHandlers) AddCarFine(c *gin.Context, plate string) {
	var fineRequest api.FineRequest
	if err := c.ShouldBindJSON(&fineRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	fine, err := fh.dao.AddCarFine(c.Request.Context(), plate, fineRequest)
	if err != nil {
		if errors.Is(err, dao.ErrCarNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "car not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to add fine"})
		return
	}

	c.JSON(http.StatusCreated, fine)
}

func (fh *FineHandlers) DeleteFines(c *gin.Context) {
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
	rolestr, ok := role.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get role"})
		return
	}
	if rolestr != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	if err := fh.dao.DeleteFines(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete all fines"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "all fines deleted successfully"})
}

func (fh *FineHandlers) GetUserFines(c *gin.Context) {
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

	fines, err := fh.dao.GetUserFines(c.Request.Context(), usernamestr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user fines"})
		return
	}
	c.JSON(http.StatusOK, fines)
}

func (fh *FineHandlers) GetFineById(c *gin.Context, id int64) {
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
	rolestr, ok := role.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get role"})
		return
	}
	if rolestr != "admin" && rolestr != "controller" {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	fine, err := fh.dao.GetFineById(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, dao.ErrFineNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "fine not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get fine"})
		return
	}

	c.JSON(http.StatusOK, fine)
}

func (fh *FineHandlers) DeleteFineById(c *gin.Context, id int64) {
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
	rolestr, ok := role.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get role"})
		return
	}
	// Check if the user has permission to delete the fine
	if rolestr != "admin" && rolestr != "controller" {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	if err := fh.dao.DeleteFineById(c.Request.Context(), id); err != nil {
		if errors.Is(err, dao.ErrFineNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "fine not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete fine"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "fine deleted successfully"})
}

func (fh *FineHandlers) PayFine(c *gin.Context, id int64) {
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
	_, ok = role.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get role"})
		return
	}

	if err := fh.dao.PayFine(c.Request.Context(), id); err != nil {
		if errors.Is(err, dao.ErrFineNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "fine not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to pay fine"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "fine paid successfully"})
}
