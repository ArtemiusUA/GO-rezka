package pages

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"go_rezka/internal/helpers"
	"go_rezka/internal/parsing"
	"go_rezka/internal/storage"
	"net/http"
	"strconv"
)

type VideoTemplateData struct {
	Video     storage.Video
	VideoUrls []storage.VideoUrl
	Parts     []storage.Part
}

func Video(w http.ResponseWriter, req *http.Request) {
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
