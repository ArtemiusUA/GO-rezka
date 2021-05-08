package main

import (
	"encoding/json"
	"github.com/gocolly/colly/v2"
	"github.com/jmoiron/sqlx/types"
	log "github.com/sirupsen/logrus"
	"go_rezka/internal/storage"
	"regexp"
	"strconv"
	"strings"
)

func main() {
	err := storage.InitDB()
	if err != nil {
		log.Fatal(err)
	}

	baseCollector := colly.NewCollector(
		colly.AllowedDomains("rezka.ag", "www.rezka.ag"),
		colly.CacheDir("./cache"),
		colly.URLFilters(
			regexp.MustCompile(`https://rezka\.ag/films/$`),
			regexp.MustCompile(`https://rezka\.ag/films/page/.+/$`),
			regexp.MustCompile(`https://rezka\.ag/films/.+/.+\.html`),
		),
	)
	videoCollector := colly.NewCollector()

	baseCollector.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		if regexp.MustCompile(`https://rezka\.ag/films/.+/.+\.html`).Match([]byte(link)) {
			videoCollector.Visit(link)
		} else {
			baseCollector.Visit(link)
		}
	})

	videoCollector.OnRequest(func(r *colly.Request) {
		log.Debugf("Visiting: %v", r.URL.String())
	})
	videoCollector.OnHTML(".b-post", func(e *colly.HTMLElement) {
		name := e.ChildText(".b-content__main .b-post__title [itemprop=name]")
		nameOrig := e.ChildText(".b-content__main [itemprop=alternativeHeadline]")
		url := e.ChildAttr("[itemprop=url]", "content")
		imageUrl := e.ChildAttr("[itemprop=image]", "src")
		genres_names := e.ChildTexts("[itemprop=genre]")
		var genres []storage.Genre
		for _, g := range genres_names {
			genres = append(genres, storage.Genre{Name: g})
		}

		description := e.ChildText(".b-post__description_text")
		rating, _ := strconv.ParseFloat(e.ChildText(".b-post__info_rates.imdb span"), 64)

		streamsContent := regexp.MustCompile(`streams\":\"((.)+?)\",`).FindSubmatch(e.Response.Body)
		if streamsContent == nil || len(streamsContent) == 0 {
			log.Warningf("No streams for: %v", url)
			return
		}
		streams := string(streamsContent[1])
		parts := strings.Split(streams, ",")
		var urls []storage.VideoUrl
		for _, p := range parts {
			firstPart := strings.Split(p, " or ")[0]
			mp4url := strings.ReplaceAll(strings.Split(p, " or ")[1], "\\", "")
			qualityPre := strings.Split(firstPart, "]")[0]
			m3u8url := strings.ReplaceAll(strings.Split(firstPart, "]")[1], "\\", "")
			quality := qualityPre[1:len(qualityPre)]
			urls = append(urls, storage.VideoUrl{quality, mp4url, m3u8url})
		}

		urlsJSONText, _ := json.Marshal(urls)

		video := storage.Video{
			0,
			name,
			nameOrig,
			url,
			imageUrl,
			description,
			rating,
			types.JSONText(urlsJSONText),
		}

		err := storage.SaveVideo(&video)
		if err != nil {
			log.Error("Error: %v", err)
			return
		}

		for _, genre := range genres {
			err = storage.SaveGenre(&genre)
			if err != nil {
				log.Error("Error: %v", err)
				return
			}
			err = storage.SaveVideoGenre(&video, &genre)
			if err != nil {
				log.Error("Error: %v", err)
				return
			}
		}

		log.Infof("Parsed video: %v", url)

	})
	videoCollector.OnError(func(r *colly.Response, err error) {
		log.Errorf("Error parsing: %v, %v", r.Request.URL, err)
	})

	err = baseCollector.Visit("https://rezka.ag/films/")
	if err != nil {
		log.Fatal(err)
	}

}
