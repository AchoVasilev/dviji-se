package server

import (
	"log"
	"net/http"
	"os"
	"server/common/middleware"
)

type ApiServer struct {
	port string
	http *http.Server
}

func initialize() *ApiServer {
	router := RegisterRoutes()

	middleware.CreateChain()

	stack := middleware.CreateChain(
		middleware.EnableCORS,
		middleware.Recovery,
	)

	port := os.Getenv("PORT")
	return &ApiServer{
		port: port,
		http: &http.Server{
			Addr:    ":" + port,
			Handler: stack(router),
		}}
}

func Run() error {
	server := initialize()

	log.Println("Starting server on port: " + server.port)

	return server.http.ListenAndServe()
}
