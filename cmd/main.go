// Copyright (c) 2025 Tiago Melo. All rights reserved.
// Use of this source code is governed by the MIT License that can be found in
// the LICENSE file.

package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jessevdk/go-flags"
	"github.com/pkg/errors"
	"github.com/tiagomelo/go-airports-service/db"
	"github.com/tiagomelo/go-airports-service/handlers"
)

type options struct {
	Port int `short:"p" long:"port" description:"server's port" required:"true"`
}

func run(port int, log *slog.Logger) error {
	ctx := context.Background()
	defer log.InfoContext(ctx, "Completed")

	// =========================================================================
	// Database support

	const sqliteDbFile = "db/airportsRestApi.db"
	db, err := db.ConnectToSqlite(sqliteDbFile)
	if err != nil {
		return errors.Wrapf(err, "opening database file %s", sqliteDbFile)
	}
	defer db.Close()

	// =========================================================================
	// API Service

	apiMux := handlers.NewApiMux(&handlers.ApiMuxConfig{
		Db:  db,
		Log: log,
	})

	// Server to service the requests against the mux.
	srv := http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: apiMux,
	}

	// Channel to listen for an interrupt or terminate signal from the OS.
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Channel to listen for errors coming from the listener.
	serverErrors := make(chan error, 1)

	// Start the service listening for api requests.
	go func() {
		log.Info(fmt.Sprintf("API listening on %s", srv.Addr))
		serverErrors <- srv.ListenAndServe()
	}()

	// Blocking main and waiting for shutdown.
	select {
	case err := <-serverErrors:
		return errors.Wrap(err, "server error")
	case sig := <-shutdown:
		log.InfoContext(ctx, fmt.Sprintf("Starting shutdown: %v", sig))

		// Give outstanding requests a deadline for completion.
		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		// Asking listener to shutdown and shed load.
		if err := srv.Shutdown(ctx); err != nil {
			srv.Close()
			return errors.Wrap(err, "could not stop server gracefully")
		}
		// Close the database connection.
		if err := db.Close(); err != nil {
			return errors.Wrap(err, "could not close database connection")
		}
	}
	return nil
}

func main() {
	var opts options
	parser := flags.NewParser(&opts, flags.Default)
	_, err := parser.Parse()
	if err != nil {
		os.Exit(1)
	}
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	if err := run(opts.Port, log); err != nil {
		log.Error("error", slog.Any("err", err))
		os.Exit(1)
	}
}
