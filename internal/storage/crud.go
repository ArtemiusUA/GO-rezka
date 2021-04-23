package storage

import "strconv"

const DEFAULT_VIDEOS_BATCH = 12

func ListGenres() (genres []Genre, err error) {
	db, err := GetDB()
	if err != nil {
		return nil, err
	}

	sql := `SELECT * FROM genres`

	err = db.Select(&genres, sql)
	if err != nil {
		return nil, err
	}

	return genres, nil
}

func ListVideos(page int, genreId int, q string) (videos []Video, err error) {
	db, err := GetDB()
	if err != nil {
		return nil, err
	}

	var params []interface{}

	sql := `SELECT DISTINCT ON (videos.id, videos.rating) videos.*
			FROM videos
			LEFT JOIN videos_genres ON videos.id = videos_genres.video_id
			WHERE 1=1 `

	if genreId != 0 {
		sql = sql + " AND videos_genres.genre_id = $1 "
		params = append(params, genreId)
	}

	if q != "" {
		sql = sql + " AND videos.name ilike $2 "
		params = append(params, q)
	}

	sql = sql + " ORDER BY videos.rating DESC LIMIT " + strconv.Itoa(DEFAULT_VIDEOS_BATCH)

	err = db.Select(&videos, sql, params...)
	if err != nil {
		return nil, err
	}

	return videos, nil
}

func GetVideosPages(genreId int, q string) (pages int, err error) {
	db, err := GetDB()
	if err != nil {
		return 0, err
	}

	var params []interface{}

	sql := `SELECT COUNT(DISTINCT videos.id) as qty
			FROM videos
			LEFT JOIN videos_genres ON videos.id = videos_genres.video_id
			WHERE 1=1 `

	if genreId != 0 {
		sql = sql + " AND videos_genres.genre_id = $1 "
		params = append(params, genreId)
	}

	if q != "" {
		sql = sql + " AND videos.name ilike $2 "
		params = append(params, q)
	}

	var videosQty int

	err = db.Get(&videosQty, sql, params...)
	if err != nil {
		return 0, err
	}

	pages = videosQty/DEFAULT_VIDEOS_BATCH + 1

	return pages, nil
}

func GetVideo(videoId int) (video Video, err error) {
	db, err := GetDB()
	if err != nil {
		return video, err
	}

	sql := `SELECT * FROM videos WHERE id = $1`

	err = db.Get(&video, sql, videoId)
	if err != nil {
		return video, err
	}

	return video, nil
}
