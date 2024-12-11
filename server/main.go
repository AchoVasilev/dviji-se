package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"server/common/api"
	"server/infrastructure/database"
	"server/infrastructure/environment"
	"server/web/ports/rest/routes"
)

func init() {
	environment.LoadEnvironmentVariables()
	database.ConnectDatabase()
	database.RunMigrations(database.Db)
}

func main() {
	mux := http.NewServeMux()

	port := os.Getenv("PORT")
	fmt.Printf("Starting server on port: %s \n", port)
	err := http.ListenAndServe(":8080", api.CheckCORS(mux))
	routes.CategoriesRoutes(mux)
	if err != nil {
		log.Fatal(err)
	}

}
