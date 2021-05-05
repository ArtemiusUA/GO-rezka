package storage

import (
	"fmt"
	"strconv"
)

const DefaultVideosBatch = 12

func SaveGenre(genre *Genre) (err error) {
	db, err := GetDB()
	if err != nil {
		return
	}

	sql := "INSERT INTO genres(name) VALUES (:name) ON CONFLICT (name) DO NOTHING"
	_, err = db.NamedExec(sql, genre)
	if err != nil {
		return
	}

	sql = "SELECT * FROM genres WHERE name = $1 LIMIT 1"
	err = db.Get(genre, sql, genre.Name)

	return
}

func ListGenres() (genres []Genre, err error) {
	db, err := GetDB()
	if err != nil {
		return
	}

	sql := `SELECT * FROM genres`

	err = db.Select(&genres, sql)
	if err != nil {
		return nil, err
	}

	return
}

func SaveVideo(video *Video) (err error) {

	db, err := GetDB()
	if err != nil {
		return
	}

	sql := `INSERT INTO videos(name, name_orig, url, image_url, description, rating, video_urls)
		 VALUES (:name, :name_orig, :url, :image_url, :description, :rating, :video_urls)
		 ON CONFLICT (name) DO 
		 UPDATE SET 
		     name_orig = EXCLUDED.name_orig, 
		     url = EXCLUDED.url, 
		     image_url = EXCLUDED.image_url, 
		     description = EXCLUDED.description, 
		     rating = EXCLUDED.rating, 
		     video_urls = EXCLUDED.video_urls`
	_, err = db.NamedExec(sql, video)
	if err != nil {
		return
	}

	sql = "SELECT * FROM videos WHERE name = $1 LIMIT 1"
	err = db.Get(video, sql, video.Name)

	return

}

func SaveVideoGenre(video *Video, genre *Genre) (err error) {
	db, err := GetDB()
	if err != nil {
		return
	}

	sql := `INSERT INTO videos_genres(video_id, genre_id) 
			VALUES ($1, $2) ON CONFLICT (video_id, genre_id) DO NOTHING`
	_, err = db.Exec(sql, video.Id, genre.Id)
	if err != nil {
		return
	}

	return nil
}

func ListVideos(page int, genreId int, q string) (videos []Video, err error) {
	db, err := GetDB()
	if err != nil {
		return
	}

	var params []interface{}

	sql := `SELECT DISTINCT ON (videos.id, videos.rating) videos.*
			FROM videos
			LEFT JOIN videos_genres ON videos.id = videos_genres.video_id
			WHERE 1=1 `

	if genreId != 0 {
		params = append(params, genreId)
		sql = sql + fmt.Sprintf(" AND videos_genres.genre_id = $%v ", len(params))
	}

	if q != "" {
		params = append(params, `%`+q+`%`)
		sql = sql + fmt.Sprintf(" AND videos.name ilike $%v ", len(params))
	}

	sql = sql + " ORDER BY videos.rating DESC LIMIT " + strconv.Itoa(DefaultVideosBatch)

	err = db.Select(&videos, sql, params...)
	if err != nil {
		return
	}

	return
}

func GetVideosPages(genreId int, q string) (pages int, err error) {
	db, err := GetDB()
	if err != nil {
		return
	}

	var params []interface{}

	sql := `SELECT COUNT(DISTINCT videos.id) as qty
			FROM videos
			LEFT JOIN videos_genres ON videos.id = videos_genres.video_id
			WHERE 1=1 `

	if genreId != 0 {
		params = append(params, genreId)
		sql = sql + fmt.Sprintf(" AND videos_genres.genre_id = $%v ", len(params))
	}

	if q != "" {
		params = append(params, `%`+q+`%`)
		sql = sql + fmt.Sprintf(" AND videos.name ilike $%v ", len(params))
	}

	var videosQty int

	err = db.Get(&videosQty, sql, params...)
	if err != nil {
		return
	}

	pages = videosQty/DefaultVideosBatch + 1

	return
}

func GetVideo(videoId int) (video Video, err error) {
	db, err := GetDB()
	if err != nil {
		return
	}

	sql := `SELECT * FROM videos WHERE id = $1`

	err = db.Get(&video, sql, videoId)
	if err != nil {
		return
	}

	return
}
