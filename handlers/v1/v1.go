// Copyright (c) 2025 Tiago Melo. All rights reserved.
// Use of this source code is governed by the MIT License that can be found in
// the LICENSE file.

package v1

import (
	"database/sql"
	"log/slog"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/tiagomelo/go-airports-service/handlers/v1/airports"
	"github.com/tiagomelo/go-airports-service/middleware"
)

// Config struct holds the database connection and logger.
type Config struct {
	Db  *sql.DB
	Log *slog.Logger
}

// Routes initializes and returns a new router with configured routes.
func Routes(c *Config) *mux.Router {
	router := mux.NewRouter()
	initializeRoutes(c.Db, router)
	router.Use(
		func(h http.Handler) http.Handler {
			return middleware.Logger(c.Log, h)
		},
		middleware.Compress,
		middleware.PanicRecovery,
	)
	return router
}

// initializeRoutes sets up the routes for airport operations.
func initializeRoutes(db *sql.DB, router *mux.Router) {
	airportsHandler := airports.NewHandlers(db)
	apiRouter := router.PathPrefix("/api/v1").Subrouter()
	apiRouter.HandleFunc("/airports", airportsHandler.HandleUpsert).Methods(http.MethodPost)
	apiRouter.HandleFunc("/nonstreaming/airports", airportsHandler.HandleNonStreamingUpsert).Methods(http.MethodPost)
}
