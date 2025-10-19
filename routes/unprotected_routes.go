package routes

import (
	"github.com/Neph-dev/MovieStreamServer/controllers"
	"github.com/gin-gonic/gin"
)

func UnprotectedRoutes(router *gin.Engine) {
	router.GET("/movies", controllers.GetMovies())
	
	router.POST("/register", controllers.RegisterUser())
	router.POST("/login", controllers.LoginUser())

}