package airports

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tiagomelo/go-airports-service/db/airports"
)

func TestHandleNonStreamingUpsert(t *testing.T) {
	testCases := []struct {
		name               string
		input              string
		mockIoReadAll      func(r io.Reader) ([]byte, error)
		mockJsonUnmarshal  func(data []byte, v any) error
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
			mockUpsertAirport: func(ctx context.Context, db *sql.DB, airport *airports.Airport) error {
				return nil
			},
			expectedOutput:     `{"message":"airports upserted"}`,
			expectedStatusCode: http.StatusOK,
		},
		{
			name: "error reading request body",
			input: `[{
				"name": "Aeroporto de Congonhas",
				"city": "São Paulo",
				"country": "Brasil",
				"iata_code": "CGH"
			}]`,
			mockIoReadAll: func(r io.Reader) ([]byte, error) {
				return nil, io.ErrUnexpectedEOF
			},
			expectedOutput:     `{"error":"failed to read request body"}`,
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name: "error unmarshalling JSON",
			input: `[{
				"name": "Aeroporto de Congonhas",
				"city": "São Paulo",
				"country": "Brasil",
				"iata_code": "CGH"
			}]`,
			mockIoReadAll: func(r io.Reader) ([]byte, error) {
				return []byte(`[{
					"name": "Aeroporto de Congonhas",
					"city": "São Paulo",
					"country": "Brasil",
					"iata_code": "CGH"
				}]`), nil
			},
			mockJsonUnmarshal: func(data []byte, v any) error {
				return io.ErrUnexpectedEOF
			},
			mockUpsertAirport: func(ctx context.Context, db *sql.DB, airport *airports.Airport) error {
				return nil
			},
			expectedOutput:     `{"error":"invalid JSON format"}`,
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name: "validation error",
			input: `[{
				"name": "Aeroporto de Congonhas",
				"city": "São Paulo",
				"country": "Brasil"
			}]`,
			mockUpsertAirport: func(ctx context.Context, db *sql.DB, airport *airports.Airport) error {
				return nil
			},
			expectedOutput:     `{"error":"[{\"field\":\"iata_code\",\"error\":\"iata_code is a required field\"}]"}`,
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
			mockUpsertAirport: func(ctx context.Context, db *sql.DB, airport *airports.Airport) error {
				return errors.New("database error")
			},
			expectedOutput:     `{"error":"error upserting airport: database error"}`,
			expectedStatusCode: http.StatusInternalServerError,
		},
	}
	originalIoReadAll := ioReadAll
	originalJsonUnmarshal := jsonUnmarshal
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				ioReadAll = originalIoReadAll
				jsonUnmarshal = originalJsonUnmarshal
			}()
			if tc.mockIoReadAll != nil {
				ioReadAll = tc.mockIoReadAll
			}
			if tc.mockJsonUnmarshal != nil {
				jsonUnmarshal = tc.mockJsonUnmarshal
			}
			upsertAirport = tc.mockUpsertAirport

			req, err := http.NewRequest(http.MethodPost, "/api/v1/nonstreaming/airports", bytes.NewBuffer([]byte(tc.input)))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			h := NewHandlers(nil)
			handler := http.HandlerFunc(h.HandleNonStreamingUpsert)
			handler.ServeHTTP(rr, req)

			require.Equal(t, tc.expectedStatusCode, rr.Code)
			require.JSONEq(t, tc.expectedOutput, rr.Body.String())
		})
	}
}
