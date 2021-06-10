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

func InitDB() error {
	db, err := GetDB()
	if err != nil {
		return nil
	}

	sql := `
		create table if not exists genres
		(
			id serial not null
				constraint genres_pkey
					primary key,
			name varchar(256)
				constraint genres_name_key
					unique
		);
			
		create table if not exists videos
		(
			id serial not null
				constraint videos_pkey
					primary key,
			name varchar(256)
				constraint videos_name_key
					unique,
			name_orig varchar(256),
			url text,
			image_url text,
			description text,
			rating double precision,
			video_urls json
		);
		
		create table if not exists videos_parts
		(
			id serial not null
				constraint videos_parts_pkey
					primary key,
			video_id integer not null
				constraint videos_parts_video_id_fkey
					references videos,
			name varchar(256),
			video_urls json,
		    unique(video_id, name) 
		);

		create table if not exists videos_genres
		(
			video_id integer not null
				constraint videos_genres_video_id_fkey
					references videos,
			genre_id integer not null
				constraint videos_genres_genre_id_fkey
					references genres,
			unique(video_id, genre_id)
		);`

	_, err = db.Exec(sql)
	return err
}
