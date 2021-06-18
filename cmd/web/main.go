package main

import (
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"go_rezka/internal/pages"
	"go_rezka/internal/storage"
	"net/http"
)

func main() {
	err := storage.InitDB()
	if err != nil {
		log.Fatal(err)
	}

	router := mux.NewRouter()
	router.HandleFunc("/", pages.Index)
	router.HandleFunc(`/{videoType:\w+}/`, pages.VideoType)
	router.HandleFunc("/videos/{id:[0-9]+}/refresh", pages.RefreshVideo)
	router.HandleFunc("/videos/{id:[0-9]+}", pages.Video)
	log.Fatal(http.ListenAndServe(":8000", router))
}
