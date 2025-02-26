package airports

import (
	"context"
	"database/sql"

	"github.com/pkg/errors"
)

type Airport struct {
	Name     string `json:"name"`
	City     string `json:"city"`
	Country  string `json:"country"`
	IataCode string `json:"iata_code"`
}

const upsertQuery = `
INSERT INTO airports (name, city, country, iata_code)
VALUES ($1, $2, $3, $4)
ON CONFLICT (iata_code) DO UPDATE
SET name = $1, city = $2, country = $3
`

func Upsert(ctx context.Context, db *sql.DB, airport *Airport) error {
	if _, err := db.ExecContext(ctx, upsertQuery,
		airport.Name,
		airport.City,
		airport.Country,
		airport.IataCode,
	); err != nil {
		return errors.Wrap(err, "upserting airport")
	}
	return nil
}
