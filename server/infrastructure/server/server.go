package server

import (
	"log/slog"
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

	stack := middleware.CreateChain(
		middleware.Recovery,
		middleware.EnableCompression,
		middleware.EnableCORS,
		middleware.PopulateRequestId,
		middleware.AppendLogger,
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

	slog.Info("Starting server on port: " + server.port)

	return server.http.ListenAndServe()
}
