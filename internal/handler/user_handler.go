package handler

import (
	"fmt"
	"log"
	"net/http"

	"github.com/huynhthanhthao/hrm_user_service/internal/dto"
	"github.com/huynhthanhthao/hrm_user_service/internal/helper"
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

	// Kiểm tra dữ liệu đầu vào
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("Error binding JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	input := dto.LoginInput(req)

	// Gọi service để đăng nhập
	loginResp, err := h.userService.Login(c.Request.Context(), c, input)

	if err != nil {
		var statusCode int = http.StatusInternalServerError
		if errMsg := err.Error(); len(errMsg) > 4 && errMsg[:4] == "HTTP" {
			fmt.Sscanf(errMsg, "HTTP %d:", &statusCode)
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Trả về response từ service
	c.JSON(http.StatusOK, gin.H{
		"access_token":  loginResp.AccessToken,
		"refresh_token": loginResp.RefreshToken,
		"user":          loginResp.User,
	})
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

func (h *UserHandler) GetUsersByIDsHandler(c *gin.Context) {
	var req struct {
		Ids []string `json:"ids"`
	}

	// Kiểm tra dữ liệu đầu vào
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	params := dto.UserParams{
		IDs: req.Ids,
		PaginationParams: dto.PaginationParams{
			Page:     1,
			PageSize: 10,
		},
	}

	// Gọi service để lấy danh sách người dùng
	users, totalCount, err := h.userService.GetUsersByIDs(c.Request.Context(), params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"users": users, "totalCount": totalCount})
}

func (h *UserHandler) RegisterHandler(c *gin.Context) {
	var req dto.RegisterDto

	// Kiểm tra dữ liệu đầu vào
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("Error binding JSON: %v", err)
		helper.Respond(c, http.StatusBadRequest, "Invalid input: "+err.Error(), nil)
		return
	}

	input := dto.RegisterInput(req)

	// Gọi service để đăng ký người dùng
	user, err := h.userService.Register(c.Request.Context(), c, input)

	if err != nil {
		log.Printf("Error registering user: %v", err)
		var statusCode int = http.StatusInternalServerError
		if errMsg := err.Error(); len(errMsg) > 4 && errMsg[:4] == "HTTP" {
			fmt.Sscanf(errMsg, "HTTP %d:", &statusCode)
		}
		helper.Respond(c, statusCode, err.Error(), nil)
		return
	}

	// Chuyển đổi dữ liệu người dùng sang response
	helper.Respond(c, http.StatusOK, "Đăng ký thành công!", user)
}
