// Copyright (c) 2025 Tiago Melo. All rights reserved.
// Use of this source code is governed by the MIT License that can be found in
// the LICENSE file.

package airports

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func TestUpsert(t *testing.T) {
	testCases := []struct {
		name          string
		input         *Airport
		mockClosure   func() *sql.DB
		expectedError error
	}{
		{
			name: "happy path",
			input: &Airport{
				Name:     "John F. Kennedy International Airport",
				City:     "New York",
				Country:  "United States",
				IataCode: "JFK",
			},
			mockClosure: func() *sql.DB {
				db, mock, err := sqlmock.New()
				require.NoError(t, err)
				mock.ExpectExec(regexp.QuoteMeta(upsertQuery)).
					WithArgs(
						"John F. Kennedy International Airport",
						"New York",
						"United States",
						"JFK",
					).WillReturnResult(sqlmock.NewResult(0, 1))
				return db
			},
		},
		{
			name: "error",
			input: &Airport{
				Name:     "John F. Kennedy International Airport",
				City:     "New York",
				Country:  "United States",
				IataCode: "JFK",
			},
			mockClosure: func() *sql.DB {
				db, mock, err := sqlmock.New()
				require.NoError(t, err)
				mock.ExpectExec(regexp.QuoteMeta(upsertQuery)).
					WithArgs(
						"John F. Kennedy International Airport",
						"New York",
						"United States",
						"JFK",
					).WillReturnError(sql.ErrConnDone)
				return db
			},
			expectedError: errors.New("upserting airport: sql: connection is already closed"),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db := tc.mockClosure()
			err := Upsert(context.TODO(), db, tc.input)
			if err != nil {
				if tc.expectedError == nil {
					t.Fatalf(`expected no error, got "%v"`, err)
				}
				require.Equal(t, tc.expectedError.Error(), err.Error())
			} else {
				if tc.expectedError != nil {
					t.Fatalf(`expected error "%v", got nil`, tc.expectedError)
				}
			}
		})
	}
}
