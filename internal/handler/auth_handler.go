package handler

import (
	"log"
	"net/http"
	"strings"

	"github.com/huynhthanhthao/hrm_user_service/internal/dto"
	"github.com/huynhthanhthao/hrm_user_service/internal/service"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) LoginHandler(c *gin.Context) {
	var req dto.LoginDto

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	input := dto.LoginInput(req)

	h.authService.Login(c.Request.Context(), c, input)
}

func (h *AuthHandler) GetMe(c *gin.Context) {
	// Lấy token từ header Authorization
	authHeader := c.GetHeader("Authorization")

	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is missing!"})
		return
	}

	// Extract Bearer token
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid Authorization header format"})
		return
	}
	tokenString := parts[1]

	h.authService.DecodeToken(c.Request.Context(), tokenString, c)
}

func (h *AuthHandler) RegisterHandler(c *gin.Context) {
	var req dto.RegisterDto

	// Kiểm tra dữ liệu đầu vào
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("Error binding JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	input := dto.RegisterInput(req)

	// Gọi service để đăng ký người dùng
	h.authService.Register(c.Request.Context(), c, input)
}
