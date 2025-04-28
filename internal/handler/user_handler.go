package handler

import (
	"net/http"
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
	type RegisterRequest struct {
		Username  string `json:"username" binding:"required,alphanum,min=3,max=20"`
		Password  string `json:"password" binding:"required,min=6,max=20"`
		FirstName string `json:"first_name" binding:"required,max=50"`
		LastName  string `json:"last_name" binding:"required,max=50"`
		Email     string `json:"email" binding:"required,email"`
		Phone     string `json:"phone" binding:"required,numeric,min=10,max=15"`
		WardCode  string `json:"ward_code" binding:"required,numeric,min=3,max=10"`
		Address   string `json:"address" binding:"required,max=200"`
		Gender    string `json:"gender" binding:"required,oneof=other female male"`
	}
	
	var req RegisterRequest
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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

	user, err := h.userService.Register(c.Request.Context(), input)
	
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

	// Trả về response
	c.JSON(http.StatusOK, gin.H{"message": "User registered successfully", "user": user})
}