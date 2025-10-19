package main

import (
	"fmt"

	controller "github.com/Neph-dev/MovieStreamServer/controllers"
	"github.com/gin-gonic/gin"
)

func main(){
	router := gin.Default()

	router.GET("/movies", controller.GetMovies())
	router.GET("/movie/:imdb_id", controller.GetMovieByImdbID())
	router.POST("/add-movie", controller.AddMovie())
	router.POST("/register", controller.RegisterUser())

	if err:= router.Run(":8080"); err != nil {
		fmt.Println("Failed to start server:", err)
	}
}