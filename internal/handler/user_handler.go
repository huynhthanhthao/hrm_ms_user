package handler

import (
	"fmt"
	"log"
	"net/http"
	"user/internal/dto"
	"user/internal/helper"
	"user/internal/service"

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