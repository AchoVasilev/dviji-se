package main

import (
	"log"
	"server/cmd/app/server"
	"server/cmd/db/database"
	"server/internal/infrastructure/environment"
)

func init() {
	environment.LoadEnvironmentVariables()
	db := database.ConnectDatabase()
	database.RunMigrations(db)
	server.Initialize(db)
}

func main() {
	if err := server.Run(); err != nil {
		log.Fatal(err)
	}
}
