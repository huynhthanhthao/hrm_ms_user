package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/huynhthanhthao/hrm_user_service/internal/dto"
	"github.com/huynhthanhthao/hrm_user_service/internal/service"
)

type UserHandler struct {
	userService service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{userService: *userService}
}

func (h *UserHandler) CreateUserHandler(c *gin.Context) {
	var req dto.CreateUserDTO

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.userService.CreateUser(c.Request.Context(), c, &req)
}

func (h *UserHandler) CreateUsersHandler(c *gin.Context) {
	var req []*dto.CreateUserDTO

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// h.userService.CreateUsers(c.Request.Context(), c, req)
}
