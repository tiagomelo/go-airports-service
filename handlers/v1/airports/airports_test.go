// Copyright (c) 2025 Tiago Melo. All rights reserved.
// Use of this source code is governed by the MIT License that can be found in
// the LICENSE file.

package airports

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tiagomelo/go-airports-service/db/airports"
)

func TestHandleUpsert(t *testing.T) {
	testCases := []struct {
		name               string
		input              string
		mockClosure        func(rc *mockResponseController)
		mockUpsertAirport  func(ctx context.Context, db *sql.DB, airport *airports.Airport) error
		expectedOutput     string
		expectedStatusCode int
	}{
		{
			name: "happy path",
			input: `[{
				"name": "Aeroporto de Congonhas",
				"city": "São Paulo",
				"country": "Brasil",
				"iata_code": "CGH"
			}]`,
			mockClosure:        func(rc *mockResponseController) {},
			mockUpsertAirport:  func(ctx context.Context, db *sql.DB, airport *airports.Airport) error { return nil },
			expectedOutput:     `{"message":"airports upserted"}`,
			expectedStatusCode: http.StatusOK,
		},
		{
			name: "missing opening [",
			input: `{
				"name": "Aeroporto de Congonhas",
				"city": "São Paulo",
				"country": "Brasil",
				"iata_code": "CGH"
			}]`,
			mockClosure:        func(rc *mockResponseController) {},
			expectedOutput:     `{"error":"invalid JSON: expected '[' at start"}`,
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name: "validation error",
			input: `[{
				"name": "Aeroporto de Congonhas",
				"city": "São Paulo",
				"country": "Brasil"
			}]`,
			mockClosure:        func(rc *mockResponseController) {},
			expectedOutput:     `{"error":"[{\"field\":\"iata_code\",\"error\":\"iata_code is a required field\"}]"}`,
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "invalid json structure",
			input:              `["name": "Aeroporto de Congonhas"]`,
			mockClosure:        func(rc *mockResponseController) {},
			expectedOutput:     `{"error":"invalid JSON airport structure"}`,
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name: "database error",
			input: `[{
				"name": "Aeroporto de Congonhas",
				"city": "São Paulo",
				"country": "Brasil",
				"iata_code": "CGH"
			}]`,
			mockClosure: func(rc *mockResponseController) {},
			mockUpsertAirport: func(ctx context.Context, db *sql.DB, airport *airports.Airport) error {
				return errors.New("database error")
			},
			expectedOutput:     `{"error":"error upserting airport: database error"}`,
			expectedStatusCode: http.StatusInternalServerError,
		},
		{
			name: "missing closing ]",
			input: `[{
				"name": "Aeroporto de Congonhas",
				"city": "São Paulo",
				"country": "Brasil",
				"iata_code": "CGH"
			}`,
			mockClosure:        func(rc *mockResponseController) {},
			mockUpsertAirport:  func(ctx context.Context, db *sql.DB, airport *airports.Airport) error { return nil },
			expectedOutput:     `{"error":"invalid JSON: expected ']' at end"}`,
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name: "flush error",
			input: `[{
				"name": "Aeroporto de Congonhas",
				"city": "São Paulo",
				"country": "Brasil",
				"iata_code": "CGH"
			}]`,
			mockClosure: func(rc *mockResponseController) {
				rc.FlushErr = errors.New("flush error")
			},
			mockUpsertAirport:  func(ctx context.Context, db *sql.DB, airport *airports.Airport) error { return nil },
			expectedOutput:     `{"error":"flush error"}`,
			expectedStatusCode: http.StatusInternalServerError,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rc := new(mockResponseController)
			tc.mockClosure(rc)
			upsertAirport = tc.mockUpsertAirport
			newHttpResponseController = func(_ http.ResponseWriter) responseController {
				return rc
			}
			req, err := http.NewRequest(http.MethodPost, "/api/v1/airports", bytes.NewBuffer([]byte(tc.input)))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			h := NewHandlers(nil)
			handler := http.HandlerFunc(h.HandleUpsert)
			handler.ServeHTTP(rr, req)

			require.Equal(t, tc.expectedStatusCode, rr.Code)
			require.JSONEq(t, tc.expectedOutput, rr.Body.String())
		})
	}
}

type mockResponseController struct {
	FlushErr error
}

func (m *mockResponseController) Flush() error {
	return m.FlushErr
}
