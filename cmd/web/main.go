package main

import (
	"fmt"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"go_rezka/internal/helpers"
	"go_rezka/internal/parsing"
	"go_rezka/internal/storage"
	"net/http"
	"os"
	"path"
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
	Parts     []storage.Part
}

func main() {
	err := storage.InitDB()
	if err != nil {
		log.Fatal(err)
	}

	router := mux.NewRouter()
	router.HandleFunc("/", index)
	router.HandleFunc("/videos/{id:[0-9]+}/refresh", refreshVideo)
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

	err = render(w, "index.gohtml", data)
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
		log.Warningf("Error parsing urls for: %v", videoId)
	}

	parts, err := storage.ListVideoParts(videoId)
	if err != nil {
		helpers.InternalError(w, err)
		return
	}

	data := VideoTemplateData{Video: video, VideoUrls: urls, Parts: parts}

	err = render(w, "video.gohtml", data)
	if err != nil {
		helpers.InternalError(w, err)
	}
}

func refreshVideo(w http.ResponseWriter, req *http.Request) {
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

	videoCollector := parsing.CreateVideoCollector()
	err = videoCollector.Visit(video.Url)
	if err != nil {
		helpers.InternalError(w, err)
	}

	http.Redirect(w, req, fmt.Sprintf("/videos/%v", video.Id), http.StatusFound)
}

func render(w http.ResponseWriter, tpl string, data interface{}) error {
	templatesPath := os.Getenv("TEMPLATES_PATH")
	if templatesPath == "" {
		cwd, _ := os.Getwd()
		templatesPath = path.Join(cwd, "templates")
	}
	templatesNames := []string{path.Join(templatesPath, "base.gohtml"),
		path.Join(templatesPath, "styles.gohtml")}
	templatesNames = append(templatesNames, path.Join(templatesPath, tpl))
	templates := template.Must(template.ParseFiles(templatesNames...))
	return templates.ExecuteTemplate(w, tpl, data)
}
