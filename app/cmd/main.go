package main

import (
	"log"
	"server/cmd/db/database"
	"server/internal/infrastructure/environment"
	"server/internal/server"
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
