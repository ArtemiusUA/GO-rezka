package storage

import (
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/jmoiron/sqlx"
	"os"
)

func GetDB() (db *sqlx.DB, err error) {
	databaseUrl := os.Getenv("DATABASE_URL")
	db, err = sqlx.Open("pgx", databaseUrl)
	if err != nil {
		return nil, err
	}
	return db, nil
}
