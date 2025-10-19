package main

import (
	"fmt"

	"github.com/Neph-dev/MovieStreamServer/routes"
	"github.com/gin-gonic/gin"
)

func main(){
	router := gin.Default()

	routes.UnprotectedRoutes(router)
	routes.ProtectedRoutes(router)
	
	if err := router.Run(":8080"); err != nil {
		fmt.Println("Failed to start server:", err)
	}
}