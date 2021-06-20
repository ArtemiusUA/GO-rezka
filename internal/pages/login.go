package pages

import (
	"github.com/ArtemiusUA/GO-rezka/internal/helpers"
	"github.com/spf13/viper"
	"net/http"
)

type LoginTemplateData struct {
	Message string
}

func Login(w http.ResponseWriter, req *http.Request) {
	var message string

	if helpers.IsAuthorized(req) {
		http.Redirect(w, req, "/", http.StatusFound)
		return
	}

	var token string
	switch req.Method {
	case "GET":
		token = req.URL.Query().Get("token")
	case "POST":
		err := req.ParseForm()
		if err != nil {
			helpers.InternalError(w, err)
			return
		}
		token = req.FormValue("token")
	}

	if token == viper.GetString("AUTH_TOKEN") {
		http.SetCookie(w, &http.Cookie{Name: "token", Value: token, HttpOnly: true})
		http.Redirect(w, req, "/", http.StatusFound)
		return
	} else if token != "" {
		message = "Invalid token"
	}

	err := helpers.Render(w, "login.gohtml", LoginTemplateData{Message: message})
	if err != nil {
		helpers.InternalError(w, err)
	}
}
