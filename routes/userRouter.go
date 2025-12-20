package routes

import (
	controller "go-authentication/controllers"
	"go-authentication/middleware"

	"github.com/gin-gonic/gin"
)

func UserRoutes(incomingRoutes *gin.Engine) {
	incomingRoutes.Use(middleware.Authenticate()) //Чтобы без токена юзера нельзя было
	incomingRoutes.GET("users", controller.GetUsers())
	incomingRoutes.GET("users/:user_id", controller.GetUser())
}
