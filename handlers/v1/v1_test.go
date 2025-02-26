// Copyright (c) 2025 Tiago Melo. All rights reserved.
// Use of this source code is governed by the MIT License that can be found in
// the LICENSE file.

package v1_test

import (
	"bytes"
	"database/sql"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tiagomelo/go-airports-service/db"
	"github.com/tiagomelo/go-airports-service/db/airports"
	"github.com/tiagomelo/go-airports-service/handlers"
)

var (
	testDb     *sql.DB
	testServer *httptest.Server
)

func TestMain(m *testing.M) {
	const sqliteDbFile = "../../db/airportsRestApiTest.db"
	var err error
	testDb, err = db.ConnectToSqlite(sqliteDbFile)
	if err != nil {
		fmt.Println("error when connecting to the test database:", err)
		os.Exit(1)
	}
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	apiMux := handlers.NewApiMux(&handlers.ApiMuxConfig{
		Db:  testDb,
		Log: log,
	})
	testServer = httptest.NewServer(apiMux)
	defer testServer.Close()
	exitVal := m.Run()
	if err := testDb.Close(); err != nil {
		fmt.Println("error when closing test database:", err)
		os.Exit(1)
	}
	if err := os.Remove(sqliteDbFile); err != nil {
		fmt.Println("error when deleting test database:", err)
		os.Exit(1)
	}
	os.Exit(exitVal)
}

func TestHandleUpsert(t *testing.T) {
	testCases := []struct {
		name             string
		inputFilePath    string
		outputFilePath   string
		expectedAirports []airports.Airport
		expectedStatus   int
	}{
		{
			name:           "insert two new airports",
			inputFilePath:  "../../testdata/airports/input/two_new_airports.json",
			outputFilePath: "../../testdata/airports/output/airports_upserted.json",
			expectedAirports: []airports.Airport{
				{IataCode: "ATL"},
				{IataCode: "ORD"},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "upsert: one new airport, two existing airports",
			inputFilePath:  "../../testdata/airports//input/one_new_two_existing_airports.json",
			outputFilePath: "../../testdata/airports/output/airports_upserted.json",
			expectedAirports: []airports.Airport{
				{IataCode: "ATL"},
				{IataCode: "ORD"},
				{IataCode: "LAX"},
			},
			expectedStatus: http.StatusOK,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			input, err := os.ReadFile(tc.inputFilePath)
			require.NoError(t, err)
			expectedOutput, err := os.ReadFile(tc.outputFilePath)
			require.NoError(t, err)

			resp, err := http.Post(testServer.URL+"/api/v1/airports", "application/json", bytes.NewBuffer([]byte(input)))
			require.NoError(t, err)
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			require.Equal(t, tc.expectedStatus, resp.StatusCode)
			require.JSONEq(t, string(expectedOutput), string(body))

			rows, err := testDb.Query("SELECT iata_code FROM airports")
			require.NoError(t, err)
			defer rows.Close()

			var createdAirports []airports.Airport
			for rows.Next() {
				var a airports.Airport
				require.NoError(t, rows.Scan(&a.IataCode))
				createdAirports = append(createdAirports, a)
			}
			require.Len(t, createdAirports, len(tc.expectedAirports))
			require.ElementsMatch(t, tc.expectedAirports, createdAirports)
			require.NoError(t, rows.Err())
		})
	}
}
