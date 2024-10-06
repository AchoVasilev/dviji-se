package main

import (
	"server/infrastructure/database"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()

	database.ConnectDatabase()
	database.RunMigrations(database.Db)

	err := router.Run(":8080")
	if err != nil {
		panic(err)
	}
}
