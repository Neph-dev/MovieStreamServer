package main

import (
	"fmt"

	controller "github.com/Neph-dev/MovieStreamServer/controllers"
	"github.com/gin-gonic/gin"
)

func main(){
	router := gin.Default()

	router.GET("/movies", controller.GetMovies())

	if err:= router.Run(":8080"); err != nil {
		fmt.Println("Failed to start server:", err)
	}
}