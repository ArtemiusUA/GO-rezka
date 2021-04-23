package helpers

import (
	"log"
	"net/http"
)

func InternalError(w http.ResponseWriter, err error) {
	message := "Internal error"
	if !IsDebug() {
		message = err.Error()
	}
	log.Println(err.Error())
	http.Error(w, message, 500)
}
