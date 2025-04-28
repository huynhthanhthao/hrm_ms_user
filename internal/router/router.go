package router

import (
	"user/ent"
	"user/internal/handler"
	"user/internal/service"

	"github.com/gin-gonic/gin"
)

func SetupRouter(client *ent.Client) *gin.Engine {
	r := gin.Default()

	userService := service.NewUserService(client)
	userHandler := handler.NewUserHandler(userService)
	
	r.POST("/register", userHandler.RegisterHandler)

	return r
}
