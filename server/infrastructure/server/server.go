package server

import (
	"log"
	"net/http"
	"os"
	"server/common/middleware"
)

func initialize() *http.Server {
	router := http.NewServeMux()

	middleware.CreateChain()
	port := os.Getenv("PORT")

	stack := middleware.CreateChain(
		middleware.ValidateBody,
		middleware.EnableCORS,
	)

	return &http.Server{
		Addr:    ":" + port,
		Handler: stack(router),
	}
}

func Run() error {
	server := initialize()

	log.Println("Starting server on port: " + server.Addr)

	return server.ListenAndServe()
}
