// Copyright (c) 2025 Tiago Melo. All rights reserved.
// Use of this source code is governed by the MIT License that can be found in
// the LICENSE file.

package airports

import (
	"bufio"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/tiagomelo/go-airports-service/db/airports"
	"github.com/tiagomelo/go-airports-service/validate"
	"github.com/tiagomelo/go-airports-service/web"
)

// UpsertAirportRequest represents a request to upsert an airport.
type UpsertAirportRequest struct {
	Name     string `json:"name" validate:"required"`
	City     string `json:"city" validate:"required"`
	Country  string `json:"country" validate:"required"`
	IataCode string `json:"iata_code" validate:"required"`
}

// ToAirport converts an upsert airport request to an airport.
func (u *UpsertAirportRequest) ToAirport() *airports.Airport {
	return &airports.Airport{
		Name:     u.Name,
		City:     u.City,
		Country:  u.Country,
		IataCode: u.IataCode,
	}
}

// UpsertAirportResponse represents a response to an upsert airport request.
type UpsertAirportResponse struct {
	Message string `json:"message"`
}

// responseController is an interface that wraps the Flush method.
type responseController interface {
	Flush() error
}

// handlerError is a custom error type that carries an HTTP status code.
type handlerError struct {
	code int
	msg  string
}

func (he handlerError) Error() string {
	return he.msg
}

// handlers struct holds a database connection.
type handlers struct {
	db *sql.DB
}

// maxBufferedReaderSize is the maximum size of the buffered reader.
const maxBufferedReaderSize = 32 * 1024

// For ease of unit testing.
var (
	// newHttpResponseController is a function that creates a new response controller.
	newHttpResponseController = func(rw http.ResponseWriter) responseController {
		return http.NewResponseController(rw)
	}
	// upsertAirport is a function that upserts an airport in the database.
	upsertAirport = airports.Upsert
)

// NewHandlers initializes a new instance of handlers with a database connection.
func NewHandlers(db *sql.DB) *handlers {
	return &handlers{
		db: db,
	}
}

// HandleUpsert handles the upsert of airports in a streaming fashion.
func (h *handlers) HandleUpsert(w http.ResponseWriter, r *http.Request) {
	ctr := newHttpResponseController(w)
	bufReader := bufio.NewReaderSize(r.Body, maxBufferedReaderSize)
	dec := json.NewDecoder(bufReader)
	// check for opening '['.
	if err := h.readExpectedToken(dec, json.Delim('[')); err != nil {
		web.RespondWithError(w, http.StatusBadRequest, "invalid JSON: expected '[' at start")
		return
	}
	// process each airport in the JSON object.
	if herr := h.processAirports(r.Context(), dec); herr != nil {
		web.RespondWithError(w, herr.code, herr.Error())
		return
	}
	// check for closing ']'.
	if err := h.readExpectedToken(dec, json.Delim(']')); err != nil {
		web.RespondWithError(w, http.StatusBadRequest, "invalid JSON: expected ']' at end")
		return
	}
	// flush response and finalize.
	if err := ctr.Flush(); err != nil {
		web.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	web.RespondAfterFlush(w, UpsertAirportResponse{Message: "airports upserted"})
}

// readExpectedToken reads the next token and verifies it matches the expected delimiter.
func (h *handlers) readExpectedToken(dec *json.Decoder, expected json.Delim) error {
	tok, err := dec.Token()
	if err != nil {
		return err
	}
	if tok != expected {
		return fmt.Errorf("unexpected token: got %v, expected %v", tok, expected)
	}
	return nil
}

// processAirport handles processing of a single airport entry.
func (h *handlers) processAirport(ctx context.Context, dec *json.Decoder) *handlerError {
	var req UpsertAirportRequest
	if err := dec.Decode(&req); err != nil {
		return &handlerError{http.StatusBadRequest, "invalid JSON airport structure"}
	}
	if err := validate.Check(req); err != nil {
		return &handlerError{http.StatusBadRequest, err.Error()}
	}
	if err := upsertAirport(ctx, h.db, req.ToAirport()); err != nil {
		return &handlerError{http.StatusInternalServerError, fmt.Sprintf("%s: %v", "error upserting airport", err)}
	}
	return nil
}

// processAirports processes all airports in the JSON array.
func (h *handlers) processAirports(ctx context.Context, dec *json.Decoder) *handlerError {
	for dec.More() {
		if herr := h.processAirport(ctx, dec); herr != nil {
			return herr
		}
	}
	return nil
}
