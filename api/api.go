package main

import (
	"github.com/ethrousseau/weblens/api/routes"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()

	router := gin.Default()

	routes.AddApiRoutes(router)
	//routes.AddUiRoutes(router)

	router.Run("127.0.0.1:4000")
}