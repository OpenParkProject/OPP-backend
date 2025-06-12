package handlers

import (
	"OPP/backend/api"
	"OPP/backend/auth"
	"OPP/backend/dao"
	"net/http"

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

func (th *TotemHandlers) GetTotemConfig(c *gin.Context, id string) {
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

	// Check if zone exists
	_, err := dao.NewZoneDao().GetZoneById(c.Request.Context(), totemRequest.ZoneId)
	if err != nil {
		if err == dao.ErrZoneNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "zone not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check if zone exists"})
		return
	}

	if err := th.dao.AddTotem(c.Request.Context(), totemRequest); err != nil {
		if err == dao.ErrTotemAlreadyExists {
			c.JSON(http.StatusConflict, gin.H{"error": "totem already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to register totem"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "totem registered successfully"})
}

func (th *TotemHandlers) GetAllTotems(c *gin.Context, params api.GetAllTotemsParams) {
	totems, err := th.dao.GetTotems(c.Request.Context(), *params.Limit, *params.Offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get totems"})
		return
	}
	c.JSON(http.StatusOK, totems)
}

func (th *TotemHandlers) DeleteTotemById(c *gin.Context, id string) {
	username, role, err := auth.GetPermissions(c)
	if err != nil {
		return
	}

	// Get totem by ID
	err, totem := th.dao.GetTotemById(c.Request.Context(), id)
	if err != nil {
		if err == dao.ErrTotemNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "totem not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get totem"})
		return
	}
	isZoneAdmin, err := NewZoneHandler().isZoneAdmin(c, totem.ZoneId, username)

	if role != "superuser" && !isZoneAdmin {
		c.JSON(http.StatusForbidden, gin.H{"forbidden": "you do not have permission to delete this totem"})
		return
	}

	err = th.dao.DeleteTotemById(c.Request.Context(), id)
	if err != nil {
		if err == dao.ErrTotemNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "totem not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete totem"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "totem deleted successfully"})
}
