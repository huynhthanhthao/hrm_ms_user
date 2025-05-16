package router

import (
	"github.com/huynhthanhthao/hrm_user_service/ent"
	"github.com/huynhthanhthao/hrm_user_service/internal/handler"
	"github.com/huynhthanhthao/hrm_user_service/internal/service"

	"github.com/gin-gonic/gin"
)

func SetupRouter(client *ent.Client, hrClients *service.HRServiceClients) *gin.Engine {
	r := gin.Default()

	authService, err := service.NewAuthService(client, hrClients)
	if err != nil {
		panic("failed to create auth service: " + err.Error())
	}
	authHandler := handler.NewAuthHandler(authService)

	// Đăng ký route
	r.POST("/login", authHandler.LoginHandler)
	r.GET("/me", authHandler.GetMe)
	r.POST("/register", authHandler.RegisterHandler)

	return r
}
