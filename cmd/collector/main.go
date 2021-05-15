package main

import (
	log "github.com/sirupsen/logrus"
	"go_rezka/internal/parsing"
	"go_rezka/internal/storage"
	"os"
)

func main() {
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
		log.Fatal(err)
	}
}
