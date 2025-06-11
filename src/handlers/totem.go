package handlers

import (
	"OPP/backend/api"
	"OPP/backend/auth"
	"OPP/backend/dao"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type TotemHandlers struct {
	dao dao.TotemDao
}

func NewTotemHandler() *TotemHandlers {
	return &TotemHandlers{
		dao: *dao.NewTotemDao(),
	}
}

func (th *TotemHandlers) GetTotemConfig(c *gin.Context, params api.GetTotemConfigParams) {
	id, err := strconv.ParseInt(params.Id, 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid totem ID"})
		return
	}
	err, totemConfig := th.dao.GetTotemById(c.Request.Context(), id)
	if err != nil {
		if err == dao.ErrTotemNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "totem not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get totem config"})
		return
	}
	c.JSON(http.StatusOK, totemConfig)
}

func (th *TotemHandlers) RegisterTotem(c *gin.Context) {
	var totemRequest api.TotemRequest
	if err := c.ShouldBindJSON(&totemRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	// Validate OTP
	if err := auth.ValidateOTP(totemRequest.Otp); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid OTP"})
		return
	}

	if err := th.dao.AddTotem(c.Request.Context(), totemRequest); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to register totem"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "totem registered successfully"})
}
