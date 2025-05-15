package handler

import (
	"net/http"

	"github.com/huynhthanhthao/hrm_user_service/internal/dto"
	"github.com/huynhthanhthao/hrm_user_service/internal/service"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

func (h *UserHandler) LoginHandler(c *gin.Context) {
	var req dto.LoginDto

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	input := dto.LoginInput(req)

	h.userService.Login(c.Request.Context(), c, input)
}

func (h *UserHandler) GetMe(c *gin.Context) {
	// Lấy token từ header Authorization
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is missing!"})
		return
	}

	// Extract Bearer token
	var token string
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		token = authHeader[7:]
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid Authorization header format!"})
		return
	}

	// Decode token để lấy thông tin người dùng
	userInfo, err := h.userService.DecodeToken(token)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"users": userInfo})
}
