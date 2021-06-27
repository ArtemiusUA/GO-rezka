package pages

import (
	"fmt"
	"github.com/ArtemiusUA/GO-rezka/internal/helpers"
	"github.com/ArtemiusUA/GO-rezka/internal/parsing"
	"github.com/ArtemiusUA/GO-rezka/internal/storage"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"net/http"
	"strconv"
)

type VideoTemplateData struct {
	Video     storage.Video
	VideoUrls []storage.VideoUrl
	Parts     []storage.Part
}

func Video(w http.ResponseWriter, req *http.Request) {
	if !helpers.IsAuthorized(req) {
		http.Redirect(w, req, "/login", http.StatusFound)
		return
	}

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
		logrus.Warningf("Error parsing urls for: %v", videoId)
	}

	parts, err := storage.ListVideoParts(videoId)
	if err != nil {
		helpers.InternalError(w, err)
		return
	}

	data := VideoTemplateData{Video: video, VideoUrls: urls, Parts: parts}

	err = helpers.Render(w, "video.gohtml", data)
	if err != nil {
		helpers.InternalError(w, err)
	}
}

func RefreshVideo(w http.ResponseWriter, req *http.Request) {
	if !helpers.IsAuthorized(req) {
		http.Redirect(w, req, "/login", http.StatusFound)
		return
	}

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
	err = videoCollector.Visit("https://" + viper.GetString("BASE_DOMAIN") + video.Url)
	if err != nil {
		log.Error("Unable to parse video: %v", err)
	}

	http.Redirect(w, req, fmt.Sprintf("/videos/%v", video.Id), http.StatusFound)
}
