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

func (h *UserHandler) RegisterHandler(c *gin.Context) {
	var req dto.RegisterRequest

	// Kiểm tra dữ liệu đầu vào
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("Error binding JSON: %v", err)
		helper.Respond(c, http.StatusBadRequest, "Invalid input: "+err.Error(), nil)
		return
	}

	input := service.RegisterInput{
		Username:  req.Username,
		Password:  req.Password,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Email:     req.Email,
		Phone:     req.Phone,
		WardCode:  req.WardCode,
		Address:   req.Address,
		Gender:    req.Gender,
		CompanyId: req.CompanyId,
	}

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

func (h *UserHandler) LoginHandler(c *gin.Context) {
	var req dto.LoginRequest

	// Kiểm tra dữ liệu đầu vào
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("Error binding JSON: %v", err)
		helper.Respond(c, http.StatusBadRequest, "Invalid input: "+err.Error(), nil)
		return
	}

	input := service.LoginInput{
		Username: req.Username,
		Password: req.Password,
	}

	// Gọi service để đăng nhập
	loginResp, err := h.userService.Login(c.Request.Context(), c, input)

	if err != nil {
		log.Printf("Error logging in: %v", err)
		var statusCode int = http.StatusInternalServerError
		if errMsg := err.Error(); len(errMsg) > 4 && errMsg[:4] == "HTTP" {
			fmt.Sscanf(errMsg, "HTTP %d:", &statusCode)
		}
		helper.Respond(c, statusCode, err.Error(), nil)
		return
	}

	helper.Respond(c, http.StatusOK, "Đăng nhập thành công!", loginResp)
}

func (h *UserHandler) GetMe(c *gin.Context) {
	// Lấy token từ header Authorization
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		helper.Respond(c, http.StatusUnauthorized, "Authorization header is missing", nil)
		return
	}

	// Tách Bearer token
	var token string
	fmt.Sscanf(authHeader, "Bearer %s", &token)
	if token == "" {
		helper.Respond(c, http.StatusUnauthorized, "Invalid token format", nil)
		return
	}

	// Decode token để lấy thông tin người dùng
	userInfo, err := h.userService.DecodeToken(token)
	if err != nil {
		helper.Respond(c, http.StatusUnauthorized, "Invalid or expired token", nil)
		return
	}

	// Trả về thông tin người dùng
	helper.Respond(c, http.StatusOK, "Thông tin người dùng", userInfo)
}
