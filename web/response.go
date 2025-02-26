// Copyright (c) 2025 Tiago Melo. All rights reserved.
// Use of this source code is governed by the MIT License that can be found in
// the LICENSE file.

package web

import (
	"encoding/json"
	"net/http"
)

// for ease of unit testing.
var (
	// marshalJson is a variable that holds the json.Marshal function.
	marshalJson = json.Marshal

	writeResponse = func(w http.ResponseWriter, response []byte) (int, error) {
		return w.Write(response)
	}
)

// RespondWithError responds a json with an error message.
func RespondWithError(w http.ResponseWriter, code int, message string) {
	Respond(w, code, map[string]string{"error": message})
}

// Respond responds a json with a payload.
func Respond(w http.ResponseWriter, code int, payload interface{}) {
	response, err := marshalJson(payload)
	if err != nil {
		errorResponse := `{"error": "internal server error"}`
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(errorResponse))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if _, err := writeResponse(w, response); err != nil {
		errorResponse := `{"error": "internal server error"}`
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(errorResponse))
	}
}

// RespondAfterFlush responds a json with a payload after flushing the response.
func RespondAfterFlush(w http.ResponseWriter, payload interface{}) {
	response, err := marshalJson(payload)
	if err != nil {
		errorResponse := `{"error": "internal server error"}`
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(errorResponse))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if _, err := writeResponse(w, response); err != nil {
		errorResponse := `{"error": "internal server error"}`
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(errorResponse))
	}
}
