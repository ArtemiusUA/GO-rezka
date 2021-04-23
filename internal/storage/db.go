package storage

import (
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/jmoiron/sqlx"
)

func GetDB() (db *sqlx.DB, err error) {
	db, err = sqlx.Open("pgx", "postgres://postgres:pass@localhost:5432/video_aggregator?sslmode=disable")
	if err != nil {
		return nil, err
	}
	return db, nil
}
