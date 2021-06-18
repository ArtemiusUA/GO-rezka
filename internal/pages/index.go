package pages

import (
	"go_rezka/internal/helpers"
	"go_rezka/internal/storage"
	"net/http"
	"strconv"
)

type IndexTemplateData struct {
	Q                string
	Page             int
	PrevPage         int
	NextPage         int
	Pages            int
	Genres           []storage.Genre
	Videos           []storage.Video
	VideoTypesTitles map[string]string
}

func Index(w http.ResponseWriter, req *http.Request) {
	q := req.URL.Query().Get("q")
	page, err := strconv.Atoi(req.URL.Query().Get("page"))
	if page <= 0 {
		page = 1
	}

	videos, err := storage.ListVideos(page, "", 0, q)
	if err != nil {
		helpers.InternalError(w, err)
		return
	}

	pages, err := storage.GetVideosPages("", 0, q)
	if err != nil {
		helpers.InternalError(w, err)
		return
	}

	data := IndexTemplateData{
		Q:                q,
		Page:             page,
		PrevPage:         page - 1,
		NextPage:         page + 1,
		Pages:            pages,
		Videos:           videos,
		VideoTypesTitles: storage.VideoTypesTitles,
	}

	err = helpers.Render(w, "index.gohtml", data)
	if err != nil {
		helpers.InternalError(w, err)
	}
}
