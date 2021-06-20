package main

import (
	"github.com/ArtemiusUA/GO-rezka/internal/helpers"
	"github.com/ArtemiusUA/GO-rezka/internal/pages"
	"github.com/ArtemiusUA/GO-rezka/internal/storage"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"net/http"
)

func main() {
	helpers.InitConfig()

	err := storage.InitDB()
	if err != nil {
		log.Fatal(err)
	}

	router := mux.NewRouter()
	router.HandleFunc("/", pages.Index)
	router.HandleFunc("/login", pages.Login)
	router.HandleFunc(`/{videoType:\w+}/`, pages.VideoType)
	router.HandleFunc("/videos/{id:[0-9]+}/refresh", pages.RefreshVideo)
	router.HandleFunc("/videos/{id:[0-9]+}", pages.Video)
	log.Fatal(http.ListenAndServe(":8000", router))
}
