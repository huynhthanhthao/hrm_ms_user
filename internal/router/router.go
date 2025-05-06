package router

import (
	"github.com/huynhthanhthao/hrm_user_service/ent"
	"github.com/huynhthanhthao/hrm_user_service/internal/handler"
	"github.com/huynhthanhthao/hrm_user_service/internal/service"

	"github.com/gin-gonic/gin"
)

func SetupRouter(client *ent.Client) *gin.Engine {
	r := gin.Default()

	userService := service.NewUserService(client)
	userHandler := handler.NewUserHandler(userService)

	r.POST("/register", userHandler.RegisterHandler)
	r.POST("/login", userHandler.LoginHandler)
	r.GET("/me", userHandler.GetMe)

	return r
}
