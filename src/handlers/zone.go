package handlers

import (
	"OPP/backend/api"
	"OPP/backend/auth"
	"OPP/backend/dao"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ZoneHandlers struct {
	dao dao.ZoneDao
}

func NewZoneHandler() *ZoneHandlers {
	return &ZoneHandlers{
		dao: *dao.NewZoneDao(),
	}
}

// TODO: move response handling outside
func (zh *ZoneHandlers) isZoneAdmin(c *gin.Context, zoneId int64, username string) (bool, error) {
	ZoneUserRole, err := zh.dao.GetZoneUserRole(c.Request.Context(), zoneId, username)
	if err != nil && !errors.Is(err, dao.ErrZoneUserRoleNotFound) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get zone user role"})
		return false, err
	}
	if errors.Is(err, dao.ErrZoneUserRoleNotFound) {
		c.JSON(http.StatusForbidden, gin.H{"error": "user not found in zone"})
		return false, nil
	}
	if ZoneUserRole.Role == "admin" {
		return true, nil
	}
	return false, nil
}

func (zh *ZoneHandlers) isZoneController(c *gin.Context, zoneId int64, username string) (bool, error) {
	ZoneUserRole, err := zh.dao.GetZoneUserRole(c.Request.Context(), zoneId, username)
	if err != nil && !errors.Is(err, dao.ErrZoneUserRoleNotFound) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get zone user role"})
		return false, err
	}
	if errors.Is(err, dao.ErrZoneUserRoleNotFound) {
		c.JSON(http.StatusForbidden, gin.H{"error": "user not found in zone"})
		return false, nil
	}
	if ZoneUserRole.Role == "controller" {
		return true, nil
	}
	return false, nil
}

func (zh *ZoneHandlers) GetZones(c *gin.Context, params api.GetZonesParams) {
	zones, err := zh.dao.GetAllZones(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get zones"})
		return
	}

	c.JSON(http.StatusOK, zones)
}

func (zh *ZoneHandlers) CreateZone(c *gin.Context) {
	username, role, err := auth.GetPermissions(c)
	if err != nil {
		return
	}
	if role != "superuser" && role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	var request api.ZoneRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	zone, err := zh.dao.CreateZone(c.Request.Context(), request)
	if err != nil {
		if errors.Is(err, dao.ErrZoneAlreadyExists) {
			c.JSON(http.StatusConflict, gin.H{"error": "zone already exists"})
			return
		}
		if errors.Is(err, dao.ErrZoneOverlap) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "zone overlaps with existing zone"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create zone"})
		return
	}

	// Automatically add the creator as an admin of the new zone
	userRole := api.ZoneUserRoleRequest{
		Username: username,
		Role:     "admin",
	}
	if _, err := zh.dao.AddUserToZone(c.Request.Context(), zone.Id, userRole, username); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to add user to zone"})
		return
	}

	c.JSON(http.StatusCreated, zone)
}

func (zh *ZoneHandlers) GetZoneById(c *gin.Context, id int64) {
	zone, err := zh.dao.GetZoneById(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, dao.ErrZoneNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "zone not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get zone"})
		return
	}

	c.JSON(http.StatusOK, zone)
}

func (zh *ZoneHandlers) UpdateZoneById(c *gin.Context, id int64) {
	username, role, err := auth.GetPermissions(c)
	if err != nil {
		return
	}
	isAdmin, err := zh.isZoneAdmin(c, id, username)
	if role != "superuser" && (!isAdmin || err != nil) {
		return
	}

	var request api.ZoneRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	zone, err := zh.dao.UpdateZone(c.Request.Context(), id, request)
	if err != nil {
		if errors.Is(err, dao.ErrZoneNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "zone not found"})
			return
		}
		if errors.Is(err, dao.ErrZoneOverlap) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "zone overlaps with existing zone"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update zone"})
		return
	}

	c.JSON(http.StatusOK, zone)
}

func (zh *ZoneHandlers) DeleteZoneById(c *gin.Context, id int64) {
	username, role, err := auth.GetPermissions(c)
	if err != nil {
		return
	}
	isAdmin, err := zh.isZoneAdmin(c, id, username)
	if role != "superuser" && (!isAdmin || err != nil) {
		return
	}

	err = zh.dao.DeleteZoneById(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, dao.ErrZoneNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "zone not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete zone"})
		return
	}

	c.Status(http.StatusNoContent)
}

func (zh *ZoneHandlers) GetZoneByLocation(c *gin.Context, params api.GetZoneByLocationParams) {
	zone, err := zh.dao.IsCoordinateInZone(c.Request.Context(), params.Lat, params.Lon)
	if err != nil {
		if errors.Is(err, dao.ErrZoneNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "no zone found at this location"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check zone location"})
		return
	}

	c.JSON(http.StatusOK, zone)
}

func (zh *ZoneHandlers) GetZoneUsers(c *gin.Context, id int64) {
	username, role, err := auth.GetPermissions(c)
	if err != nil {
		return
	}

	_, err = zh.dao.GetZoneById(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, dao.ErrZoneNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "zone not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check if zone exists"})
		return
	}

	isAdmin, err := zh.isZoneAdmin(c, id, username)
	if role != "superuser" && (!isAdmin || err != nil) {
		return
	}

	roles, err := zh.dao.GetZoneUserRoles(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get zone users"})
		return
	}

	if len(roles) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "no users found for this zone"})
		return
	}

	c.JSON(http.StatusOK, roles)
}

func (zh *ZoneHandlers) AddZoneUserRole(c *gin.Context, id int64) {
	username, role, err := auth.GetPermissions(c)
	if err != nil {
		return
	}

	_, err = zh.dao.GetZoneById(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, dao.ErrZoneNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "zone not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check if zone exists"})
		return
	}

	var request api.ZoneUserRoleRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if request.Role != "admin" && request.Role != "controller" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "role must be either 'admin' or 'controller'"})
		return
	}

	// Check permissions
	if request.Role == "admin" {
		if role != "superuser" {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
	} else if request.Role == "controller" {
		isAdmin, err := zh.isZoneAdmin(c, id, username)
		if role != "superuser" && (!isAdmin || err != nil) {
			return
		}
	}

	userRole, err := zh.dao.AddUserToZone(c.Request.Context(), id, request, username)
	if err != nil {
		if errors.Is(err, dao.ErrZoneUserRoleAlreadyExists) {
			c.JSON(http.StatusConflict, gin.H{"error": "user already in zone"})
			return
		}
		if errors.Is(err, dao.ErrZoneNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "zone not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to add user to zone"})
		return
	}

	c.JSON(http.StatusCreated, userRole)
}

func (zh *ZoneHandlers) RemoveZoneUserRole(c *gin.Context, id int64, username string) {
	username, role, err := auth.GetPermissions(c)
	if err != nil {
		return
	}
	isAdmin, err := zh.isZoneAdmin(c, id, username)
	if role != "superuser" && (!isAdmin || err != nil) {
		return
	}

	_, err = zh.dao.GetZoneById(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, dao.ErrZoneNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "zone not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check if zone exists"})
		return
	}

	err = zh.dao.RemoveUserFromZone(c.Request.Context(), id, username)
	if err != nil {
		if errors.Is(err, dao.ErrZoneNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "zone role not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to remove user from zone"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "user removed from zone successfully"})
}

func (zh *ZoneHandlers) GetUserZones(c *gin.Context) {
	username, _, err := auth.GetPermissions(c)
	if err != nil {
		return
	}

	zones, err := zh.dao.GetUserZones(c.Request.Context(), username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user zones"})
		return
	}

	c.JSON(http.StatusOK, zones)
}

func (zh *ZoneHandlers) GetUserZonesByUsername(c *gin.Context, username string) {
	zones, err := zh.dao.GetUserZones(c.Request.Context(), username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user zones"})
		return
	}
	if len(zones) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "no zones found for this user"})
		return
	}
	c.JSON(http.StatusOK, zones)
}
