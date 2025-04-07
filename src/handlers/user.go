package handlers

import (
	"OPP/backend/api"
	"OPP/backend/dao"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type UserHandlers struct {
	dao dao.UserDao
}

func NewUserHandler() *UserHandlers {
	return &UserHandlers{
		dao: *dao.NewUserDao(),
	}
}

func (uh *UserHandlers) GetUsers(c *gin.Context, params api.GetUsersParams) {
	users := uh.dao.GetUsers(c.Request.Context(), params.Limit, params.Offset)
	c.JSON(http.StatusOK, users)
}

func (uh *UserHandlers) GetUser(c *gin.Context, username string) {
	user, err := uh.dao.GetUser(c.Request.Context(), username)
	if err != nil {
		if errors.Is(err, dao.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user"})
		return
	}
	c.JSON(http.StatusOK, user)
}

func (uh *UserHandlers) AddUser(c *gin.Context) {
	var user api.UserRequest
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if err := uh.dao.AddUser(c.Request.Context(), user); err != nil {
		if err == dao.ErrUserAlreadyExists {
			c.JSON(http.StatusConflict, gin.H{"error": "user already exists"})
			return
		}
		if errors.Is(err, dao.ErrInvalidUser) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user data"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to add user"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "user added successfully"})
}

func (uh *UserHandlers) DeleteUsers(c *gin.Context) {
	if err := uh.dao.DeleteAllUsers(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete all users"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "all users deleted successfully"})
}

func (uh *UserHandlers) DeleteUser(c *gin.Context, username string) {
	if err := uh.dao.DeleteUser(c.Request.Context(), username); err != nil {
		if err == dao.ErrUserNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete user"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "user deleted successfully"})
}

func (uh *UserHandlers) UpdateUser(c *gin.Context, username string) {
	var userRequest api.UserRequest
	if err := c.ShouldBindJSON(&userRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if err := uh.dao.UpdateUser(c.Request.Context(), username, userRequest); err != nil {
		if err == dao.ErrUserNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		if err == dao.ErrUserAlreadyExists {
			c.JSON(http.StatusConflict, gin.H{"error": "email already in use"})
			return
		}
		if errors.Is(err, dao.ErrInvalidUser) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user data"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "user updated successfully"})
}
