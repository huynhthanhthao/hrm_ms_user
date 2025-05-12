package router

import (
	"github.com/huynhthanhthao/hrm_user_service/ent"
	"github.com/huynhthanhthao/hrm_user_service/internal/handler"
	"github.com/huynhthanhthao/hrm_user_service/internal/service"

	"github.com/gin-gonic/gin"
)

func SetupRouter(client *ent.Client) *gin.Engine {
	r := gin.Default()

	userService, err := service.NewUserService(client)
	if err != nil {
		panic(err)
	}
	userHandler := handler.NewUserHandler(userService)

	r.POST("/register", userHandler.RegisterHandler)
	r.POST("/login", userHandler.LoginHandler)
	r.GET("/me", userHandler.GetMe)
	r.GET("/users-by-ids", userHandler.GetUsersByIDsHandler)

	return r
}
