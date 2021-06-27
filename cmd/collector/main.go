package main

import (
	"github.com/ArtemiusUA/GO-rezka/internal/helpers"
	"github.com/ArtemiusUA/GO-rezka/internal/parsing"
	"github.com/ArtemiusUA/GO-rezka/internal/storage"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"os"
)

func main() {
	helpers.InitConfig()

	err := storage.InitDB()
	if err != nil {
		log.Fatal(err)
	}

	args := os.Args[1:]
	if len(args) > 0 {
		videoCollector := parsing.CreateVideoCollector()
		for _, arg := range args {
			err := videoCollector.Visit(arg)
			if err != nil {
				log.Error("Error: %v", err)
			}
		}
		return
	}

	baseDomain := viper.GetString("BASE_DOMAIN")
	baseCollector := parsing.CreateBaseCollector()
	err = baseCollector.Visit("https://" + baseDomain + "/films/")
	if err != nil {
		log.Error(err)
	}
	err = baseCollector.Visit("https://" + baseDomain + "/cartoons/")
	if err != nil {
		log.Error(err)
	}
	err = baseCollector.Visit("https://" + baseDomain + "/series/")
	if err != nil {
		log.Error(err)
	}
	err = baseCollector.Visit("https://" + baseDomain + "/animation/")
	if err != nil {
		log.Error(err)
	}
}
