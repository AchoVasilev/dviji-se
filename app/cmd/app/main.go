package main

import (
	"log"
	"server/internal/http/server"
	"server/internal/infrastructure/database"
	"server/internal/infrastructure/environment"
)

func init() {
	environment.LoadEnvironmentVariables()
	database.ConnectDatabase()
	database.RunMigrations(database.Db)
}

func main() {
	if err := server.Run(); err != nil {
		log.Fatal(err)
	}
}
