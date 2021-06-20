package helpers

import (
	"github.com/spf13/viper"
	"net/http"
)

func IsAuthorized(req *http.Request) bool {
	authToken := viper.GetString("AUTH_TOKEN")
	cookie, err := req.Cookie("token")
	if authToken == "" || err == nil && cookie.Value == authToken {
		return true
	}
	return false
}
