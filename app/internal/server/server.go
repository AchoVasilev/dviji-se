package server

import (
	"context"
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
	middleware.InitLogger()

	router := routes.RegisterRoutes(db)

	stack := middleware.CreateChain(
		middleware.Recovery,
		middleware.LimitRequestBody,
		middleware.EnableCompression,
		middleware.EnableCORS,
		// PopulateRequestId first so every log line below carries the id, and
		// CheckAuth before the CSRF pair because CSRF tokens are bound to the
		// authenticated identity.
		middleware.PopulateRequestId,
		middleware.CheckAuth,
		middleware.CSRFValidate,
		middleware.CSRFCookie,
		middleware.ContentType,
		middleware.SecurityHeaders,
		middleware.ContentSecurityPolicy,
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

func Shutdown(ctx context.Context) error {
	slog.Info("Shutting down server...")
	return api.http.Shutdown(ctx)
}
