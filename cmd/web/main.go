package main

import (
	"crypto/tls"
	"fmt"
	"github.com/ArtemiusUA/GO-rezka/internal/helpers"
	"github.com/ArtemiusUA/GO-rezka/internal/pages"
	"github.com/ArtemiusUA/GO-rezka/internal/storage"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"golang.org/x/crypto/acme/autocert"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
)

func main() {
	helpers.InitConfig()

	err := storage.InitDB()
	if err != nil {
		log.Fatal(err)
	}

	router := mux.NewRouter()
	router.HandleFunc("/", pages.Index)
	router.HandleFunc("/login", pages.Login)
	router.HandleFunc(`/{videoType:\w+}/`, pages.VideoType)
	router.HandleFunc("/videos/{id:[0-9]+}/refresh", pages.RefreshVideo)
	router.HandleFunc("/videos/{id:[0-9]+}", pages.Video)

	if viper.GetBool("HTTPS") {
		serveHTTPS()
	} else {
		log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", viper.GetInt("PORT")), router))
	}

}

func serveHTTPS() {

	domains := viper.GetStringSlice("DOMAINS")
	certManager := autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(domains...),
	}

	dir := cacheDir()
	if dir != "" {
		certManager.Cache = autocert.DirCache(dir)
	}

	server := &http.Server{
		Addr: ":https",
		TLSConfig: &tls.Config{
			GetCertificate:     certManager.GetCertificate,
			InsecureSkipVerify: true,
		},
	}

	log.Printf("Serving http/https for domains: %+v", domains)
	go func() {
		h := certManager.HTTPHandler(nil)
		log.Fatal(http.ListenAndServe(":http", h))
	}()

	log.Fatal(server.ListenAndServeTLS("", ""))
}

func cacheDir() (dir string) {
	if u, _ := user.Current(); u != nil {
		dir = filepath.Join(os.TempDir(), "cache-golang-autocert-"+u.Username)
		if err := os.MkdirAll(dir, 0700); err == nil {
			return dir
		}
	}
	return ""
}
