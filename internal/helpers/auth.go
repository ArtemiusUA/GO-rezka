package helpers

import (
	"net/http"
	"os"
)

func IsAuthorized(req *http.Request) bool {
	authToken := GetAuthToken()
	cookie, err := req.Cookie("token")
	if authToken == "" || err == nil && cookie.Value == authToken {
		return true
	}
	return false
}

func GetAuthToken() string {
	return os.Getenv("AUTH_TOKEN")
}
