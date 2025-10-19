package routes

import (
	"github.com/Neph-dev/MovieStreamServer/controllers"
	"github.com/Neph-dev/MovieStreamServer/middleware"
	"github.com/gin-gonic/gin"
)

func ProtectedRoutes(router *gin.Engine) {
	router.Use(middleware.AuthMiddleware())

	router.PUT("/add-movie", controllers.AddMovie())
	router.GET("/review/:imdb_id", controllers.AdminReviewUpdate())

	router.GET("/movie/:imdb_id", controllers.GetMovieByImdbID())
}