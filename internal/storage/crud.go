package storage

import (
	"fmt"
)

const DefaultVideosBatch = 12

func SaveGenre(genre *Genre) (err error) {
	db, err := GetDB()
	if err != nil {
		return
	}

	sql := "INSERT INTO genres(type, name) VALUES (:type, :name) ON CONFLICT (type, name) DO NOTHING"
	_, err = db.NamedExec(sql, genre)
	if err != nil {
		return
	}

	sql = "SELECT * FROM genres WHERE type = $1 and name = $2 LIMIT 1"
	err = db.Get(genre, sql, genre.Type, genre.Name)

	return
}

func ListGenres(videoType string) (genres []Genre, err error) {
	db, err := GetDB()
	if err != nil {
		return
	}

	sql := `SELECT * FROM genres WHERE 1=1 `

	var params []interface{}
	if videoType != "" {
		params = append(params, videoType)
		sql = sql + fmt.Sprintf(" AND genres.type = $%v ", len(params))
	}

	err = db.Select(&genres, sql, params...)
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

func SaveVideoPart(video *Video, part *Part) (err error) {
	db, err := GetDB()
	if err != nil {
		return
	}

	sql := `INSERT INTO videos_parts(video_id, name, video_urls, season_id, episode_id) 
			VALUES ($1, $2, $3, $4, $5) ON CONFLICT (video_id, name) DO 
			UPDATE SET 
				video_urls = EXCLUDED.video_urls
			`
	_, err = db.Exec(sql, video.Id, part.Name, part.Video_urls, part.Season_id, part.Episode_id)
	if err != nil {
		return
	}

	return nil
}

func ListVideos(page int, videoType string, genreId int, q string) (videos []Video, err error) {
	db, err := GetDB()
	if err != nil {
		return
	}

	var params []interface{}

	sql := `SELECT DISTINCT ON (videos.id, videos.rating) videos.*
			FROM videos
			LEFT JOIN videos_genres ON videos.id = videos_genres.video_id
			LEFT JOIN genres ON genres.id = videos_genres.genre_id
			WHERE 1=1 `

	if videoType != "" {
		params = append(params, videoType)
		sql = sql + fmt.Sprintf(" AND genres.type = $%v ", len(params))
	}

	if genreId != 0 {
		params = append(params, genreId)
		sql = sql + fmt.Sprintf(" AND videos_genres.genre_id = $%v ", len(params))
	}

	if q != "" {
		params = append(params, `%`+q+`%`)
		sql = sql + fmt.Sprintf(" AND videos.name ilike $%v ", len(params))
	}

	sql = sql + " ORDER BY videos.rating DESC "

	offset := page*DefaultVideosBatch - DefaultVideosBatch

	sql = sql + fmt.Sprintf(" LIMIT  %v OFFSET %v", DefaultVideosBatch, offset)

	err = db.Select(&videos, sql, params...)
	if err != nil {
		return
	}

	return
}

func GetVideosPages(videoType string, genreId int, q string) (pages int, err error) {
	db, err := GetDB()
	if err != nil {
		return
	}

	var params []interface{}

	sql := `SELECT COUNT(DISTINCT videos.id) as qty
			FROM videos
			LEFT JOIN videos_genres ON videos.id = videos_genres.video_id
			LEFT JOIN genres ON genres.id = videos_genres.genre_id
			WHERE 1=1 `

	if videoType != "" {
		params = append(params, videoType)
		sql = sql + fmt.Sprintf(" AND genres.type = $%v ", len(params))
	}

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

func ListVideoParts(videoId int) (parts []Part, err error) {
	db, err := GetDB()
	if err != nil {
		return
	}

	sql := `SELECT id, name, video_urls 
			FROM videos_parts 
			WHERE video_id = $1 
			ORDER BY season_id, episode_id, name`

	err = db.Select(&parts, sql, videoId)
	if err != nil {
		return
	}

	return
}
