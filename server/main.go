package main

import (
	"log"
	"os"
	"server/infrastructure/database"
	"server/infrastructure/environment"
	"server/web/ports/rest/routes"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func init() {
	environment.LoadEnvironmentVariables()
	database.ConnectDatabase()
	database.RunMigrations(database.Db)
}

func main() {
	router := gin.Default()
	router.Use(cors.Default())
	apiV1 := router.Group("/v1")

	routes.CategoriesRoutes(apiV1)

	err := router.Run(":" + os.Getenv("PORT"))
	if err != nil {
		log.Fatal(err)
	}
}
