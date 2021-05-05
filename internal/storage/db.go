package storage

import (
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/jmoiron/sqlx"
	"os"
)

var db *sqlx.DB
var err error

func GetDB() (*sqlx.DB, error) {
	if db != nil && err == nil {
		return db, err
	}
	databaseUrl := os.Getenv("DATABASE_URL")
	db, err = sqlx.Open("pgx", databaseUrl)
	if err != nil {
		return nil, err
	}
	return db, nil
}
