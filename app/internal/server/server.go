package server

import (
	"database/sql"
	"log/slog"
	"net/http"
	"server/internal/config"
	"server/internal/http/middleware"
	"server/internal/http/routes"
)

type ApiServer struct {
	port string
	http *http.Server
}

var api *ApiServer

func Initialize(db *sql.DB) {
	router := routes.RegisterRoutes(db)

	stack := middleware.CreateChain(
		middleware.Recovery,
		middleware.EnableCompression,
		middleware.EnableCORS,
		middleware.CSRFValidate,
		middleware.CSRFCookie,
		middleware.ContentType,
		middleware.CheckAuth,
		middleware.SecurityHeaders,
		middleware.ContentSecurityPolicy,
		middleware.PopulateRequestId,
		middleware.AppendLogger,
	)

	port := config.Port()
	api = &ApiServer{
		port: port,
		http: &http.Server{
			Addr:    ":" + port,
			Handler: stack(router),
		},
	}
}

func Run() error {
	slog.Info("Starting server on port: " + api.port)

	return api.http.ListenAndServe()
}
