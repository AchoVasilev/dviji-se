package main

import (
	"log"
	"server/infrastructure/database"
	"server/infrastructure/environment"
	"server/infrastructure/server"
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
