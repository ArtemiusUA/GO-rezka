package main

import (
	"github.com/ArtemiusUA/GO-rezka/internal/helpers"
	"github.com/ArtemiusUA/GO-rezka/internal/parsing"
	"github.com/ArtemiusUA/GO-rezka/internal/storage"
	log "github.com/sirupsen/logrus"
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

	baseCollector := parsing.CreateBaseCollector()
	err = baseCollector.Visit("https://rezka.ag/films/")
	if err != nil {
		log.Error(err)
	}
	err = baseCollector.Visit("https://rezka.ag/cartoons/")
	if err != nil {
		log.Error(err)
	}
	err = baseCollector.Visit("https://rezka.ag/series/")
	if err != nil {
		log.Error(err)
	}
	err = baseCollector.Visit("https://rezka.ag/animation/")
	if err != nil {
		log.Error(err)
	}
}
