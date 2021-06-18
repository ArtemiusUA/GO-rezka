package pages

import (
	"github.com/ArtemiusUA/GO-rezka/internal/helpers"
	"github.com/ArtemiusUA/GO-rezka/internal/storage"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
)

type VideoTypeTemplateData struct {
	VideoType      string
	GenreId        int
	Q              string
	Page           int
	PrevPage       int
	NextPage       int
	Pages          int
	Genres         []storage.Genre
	Videos         []storage.Video
	VideoTypeTitle string
}

func VideoType(w http.ResponseWriter, req *http.Request) {
	if !helpers.IsAuthorized(req) {
		http.Redirect(w, req, "/login", http.StatusFound)
		return
	}

	vars := mux.Vars(req)
	videoType := vars["videoType"]

	genreId, err := strconv.Atoi(req.URL.Query().Get("genre_id"))
	q := req.URL.Query().Get("q")
	page, err := strconv.Atoi(req.URL.Query().Get("page"))
	if page <= 0 {
		page = 1
	}

	genres, err := storage.ListGenres(videoType)
	if err != nil {
		helpers.InternalError(w, err)
		return
	}

	videos, err := storage.ListVideos(page, videoType, genreId, q)
	if err != nil {
		helpers.InternalError(w, err)
		return
	}

	pages, err := storage.GetVideosPages(videoType, genreId, q)
	if err != nil {
		helpers.InternalError(w, err)
		return
	}

	data := VideoTypeTemplateData{
		VideoType:      videoType,
		GenreId:        genreId,
		Q:              q,
		Page:           page,
		PrevPage:       page - 1,
		NextPage:       page + 1,
		Pages:          pages,
		Genres:         genres,
		Videos:         videos,
		VideoTypeTitle: storage.VideoTypesTitles[videoType],
	}

	err = helpers.Render(w, "type.gohtml", data)
	if err != nil {
		helpers.InternalError(w, err)
	}
}
