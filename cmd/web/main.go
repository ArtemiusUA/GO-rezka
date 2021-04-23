package main

import (
	"github.com/gorilla/mux"
	"go_rezka/internal/helpers"
	"go_rezka/internal/storage"
	"log"
	"net/http"
	"strconv"
	"text/template"
)

type IndexTemplateData struct {
	GenreId  int
	Q        string
	Page     int
	PrevPage int
	NextPage int
	Pages    int
	Genres   []storage.Genre
	Videos   []storage.Video
}

type VideoTemplateData struct {
	Video     storage.Video
	VideoUrls []storage.VideoUrl
}

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/", index)
	router.HandleFunc("/videos/{id:[0-9]+}", video)
	log.Fatal(http.ListenAndServe(":8000", router))
}

func index(w http.ResponseWriter, req *http.Request) {
	genreId, err := strconv.Atoi(req.URL.Query().Get("genre_id"))
	q := req.URL.Query().Get("q")
	page, err := strconv.Atoi(req.URL.Query().Get("page"))
	if page <= 0 {
		page = 1
	}

	genres, err := storage.ListGenres()
	if err != nil {
		helpers.InternalError(w, err)
		return
	}

	videos, err := storage.ListVideos(page, genreId, q)
	if err != nil {
		helpers.InternalError(w, err)
		return
	}

	pages, err := storage.GetVideosPages(genreId, q)
	if err != nil {
		helpers.InternalError(w, err)
		return
	}

	data := IndexTemplateData{
		GenreId:  genreId,
		Q:        q,
		Page:     page,
		PrevPage: page - 1,
		NextPage: page + 1,
		Pages:    pages,
		Genres:   genres,
		Videos:   videos,
	}

	err = render(w, "index.html", data)
	if err != nil {
		helpers.InternalError(w, err)
	}
}

func video(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	videoId, err := strconv.Atoi(vars["id"])
	if err != nil {
		helpers.InternalError(w, err)
		return
	}

	video, err := storage.GetVideo(videoId)
	if err != nil {
		helpers.InternalError(w, err)
		return
	}

	urls, err := video.GetUrls()
	if err != nil {
		log.Printf("Error parsing urls for: %v", videoId)
	}
	data := VideoTemplateData{Video: video, VideoUrls: urls}

	err = render(w, "video.html", data)
	if err != nil {
		helpers.InternalError(w, err)
	}
}

func render(w http.ResponseWriter, tpl string, data interface{}) error {
	templatesNames := []string{"cmd/web/templates/base.html", "cmd/web/templates/styles.html"}
	templatesNames = append(templatesNames, "cmd/web/templates/"+tpl)
	templates := template.Must(template.ParseFiles(templatesNames...))
	return templates.ExecuteTemplate(w, tpl, data)
}
