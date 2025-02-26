// Copyright (c) 2025 Tiago Melo. All rights reserved.
// Use of this source code is governed by the MIT License that can be found in
// the LICENSE file.

package db

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
)

// For ease of unit testing.
var sqlOpen = sql.Open

// ConnectToSqlite establishes a connection to a SQLite database.
func ConnectToSqlite(sqliteFilePath string) (*sql.DB, error) {
	// SQLite DSN (Data Source Name) with specific tuning options:
	// - _journal=WAL        - Enables Write-Ahead Logging (WAL) mode to improve write performance and concurrency.
	// - _synchronous=NORMAL - Reduces the number of sync operations to disk, balancing speed and durability.
	// - _cache=private      - Ensures each database connection has its own cache, preventing conflicts in multi-connection scenarios.
	// - _busy_timeout=5000  - If the database is locked, wait up to 5000ms before failing, improving robustness under contention.
	dsn := sqliteFilePath + "?_journal=WAL&_synchronous=NORMAL&_cache=private&_busy_timeout=5000"
	db, err := sqlOpen("sqlite3", dsn)
	if err != nil {
		return nil, errors.Wrapf(err, "opening sqlite file %s", sqliteFilePath)
	}
	return db, nil
}
