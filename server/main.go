package main

import (
	"fmt"
	"log"
	"net/http"
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
	mux := http.NewServeMux()
	router := gin.Default()
	router.Use(cors.Default())

	port := os.Getenv("PORT")
	fmt.Printf("Starting server on port: %s \n", port)
	err := http.ListenAndServe(":8080", mux)
	apiV1 := router.Group("/v1")
	routes.CategoriesRoutes(apiV1)
	if err != nil {
		log.Fatal(err)
	}

}
