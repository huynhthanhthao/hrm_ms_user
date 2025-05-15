package router

import (
	"github.com/huynhthanhthao/hrm_user_service/ent"
	"github.com/huynhthanhthao/hrm_user_service/internal/handler"
	"github.com/huynhthanhthao/hrm_user_service/internal/service"

	"github.com/gin-gonic/gin"
)

func SetupRouter(client *ent.Client, userService *service.UserService) *gin.Engine {
	r := gin.Default()

	userHandler := handler.NewUserHandler(userService)

	r.POST("/login", userHandler.LoginHandler)
	r.GET("/me", userHandler.GetMe)

	return r
}
