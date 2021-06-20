package helpers

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"net/http"
)

func InternalError(w http.ResponseWriter, err error) {
	message := "Internal error"
	if viper.GetBool("DEBUG") {
		message = err.Error()
	}
	log.Error(err.Error())
	http.Error(w, message, 500)
}
