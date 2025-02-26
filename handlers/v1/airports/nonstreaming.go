// Copyright (c) 2025 Tiago Melo. All rights reserved.
// Use of this source code is governed by the MIT License that can be found in
// the LICENSE file.

package airports

import (
	"encoding/json"
	"io"
	"net/http"
	"runtime"
	"runtime/debug"

	"github.com/pkg/errors"
	"github.com/tiagomelo/go-airports-service/validate"
	"github.com/tiagomelo/go-airports-service/web"
)

// for ease of unit testing.
var (
	ioReadAll     = io.ReadAll
	jsonUnmarshal = json.Unmarshal
)

// HandleNonStreamingUpsert handles the upsert of airports by reading the entire JSON array into memory.
func (h *handlers) HandleNonStreamingUpsert(w http.ResponseWriter, r *http.Request) {
	// read full request body into memory.
	body, err := ioReadAll(r.Body)
	if err != nil {
		web.RespondWithError(w, http.StatusBadRequest, "failed to read request body")
		return
	}
	var airportsToBeUpserted []UpsertAirportRequest
	if err := jsonUnmarshal(body, &airportsToBeUpserted); err != nil {
		web.RespondWithError(w, http.StatusBadRequest, "invalid JSON format")
		return
	}
	for _, request := range airportsToBeUpserted {
		if err := validate.Check(request); err != nil {
			web.RespondWithError(w, http.StatusBadRequest, err.Error())
			return
		}
		if err := upsertAirport(r.Context(), h.db, request.ToAirport()); err != nil {
			web.RespondWithError(w, http.StatusInternalServerError, errors.Wrap(err, "error upserting airport").Error())
			return
		}
	}
	web.Respond(w, http.StatusOK, UpsertAirportResponse{Message: "airports upserted"})

	// manually trigger garbage collection to free up memory.
	runtime.GC()
	debug.FreeOSMemory()
}
