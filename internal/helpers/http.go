package helpers

import (
	log "github.com/sirupsen/logrus"
	"net/http"
)

func InternalError(w http.ResponseWriter, err error) {
	message := "Internal error"
	if IsDebug() {
		message = err.Error()
	}
	log.Error(err.Error())
	http.Error(w, message, 500)
}
