package parsing

import (
	"encoding/json"
	"fmt"
	"github.com/gocolly/colly/v2"
	"github.com/jmoiron/sqlx/types"
	log "github.com/sirupsen/logrus"
	"go_rezka/internal/storage"
	"io/ioutil"
	"net/http"
	urlPackage "net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

func CreateBaseCollector() *colly.Collector {
	videoCollector := CreateVideoCollector()

	baseCollector := colly.NewCollector(
		colly.AllowedDomains("rezka.ag", "www.rezka.ag"),
		colly.CacheDir("./cache"),
		colly.URLFilters(
			regexp.MustCompile(`https://rezka\.ag/films/$`),
			regexp.MustCompile(`https://rezka\.ag/films/page/.+/$`),
			regexp.MustCompile(`https://rezka\.ag/films/.+/.+\.html`),
		),
	)

	baseCollector.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		if regexp.MustCompile(`https://rezka\.ag/films/.+/.+\.html`).Match([]byte(link)) {
			videoCollector.Visit(link)
		} else {
			baseCollector.Visit(link)
		}
	})

	return baseCollector
}

func CreateVideoCollector() *colly.Collector {
	videoCollector := colly.NewCollector()

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
		urls := parseUrls(streams)

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

		waitGroup := sync.WaitGroup{}
		e.ForEach("#translators-list li", func(i int, e *colly.HTMLElement) {
			waitGroup.Add(1)
			go parsePart(e, url, video, &waitGroup)
		})
		waitGroup.Wait()

	})
	videoCollector.OnError(func(r *colly.Response, err error) {
		log.Errorf("Error parsing: %v, %v", r.Request.URL, err)
	})

	return videoCollector
}

func parsePart(e *colly.HTMLElement, url string, video storage.Video, group *sync.WaitGroup) {
	defer group.Done()
	id := e.Attr("data-id")
	title := e.Attr("title")
	translatorId := e.Attr("data-translator_id")
	isCamrip := e.Attr("data-camrip")
	isAds := e.Attr("data-ads")
	isDirector := e.Attr("data-director")
	action := "get_movie"
	partUrl := fmt.Sprint("https://rezka.ag/ajax/get_cdn_series/?t=",
		time.Now().UnixNano()/int64(time.Millisecond))
	resp, err := http.PostForm(
		partUrl,
		urlPackage.Values{
			"id":            {id},
			"translator_id": {translatorId},
			"is_director":   {isDirector},
			"is_camrip":     {isCamrip},
			"is_ads":        {isAds},
			"action":        {action},
		},
	)
	if err != nil {
		log.Errorf("Error parsing part: %v, %v", partUrl, err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("Error parsing part: %v, %v", partUrl, err)
	}
	streamsContent := regexp.MustCompile(`url\":\"((.)+?)\",`).FindSubmatch(body)
	if streamsContent == nil || len(streamsContent) == 0 {
		log.Warningf("No streams for: %v", url)
		return
	}
	streams := string(streamsContent[1])
	urls := parseUrls(streams)
	urlsJSONText, _ := json.Marshal(urls)
	part := storage.Part{
		0,
		title,
		types.JSONText(urlsJSONText),
	}
	err = storage.SaveVideoPart(&video, &part)
	if err != nil {
		log.Error("Error: %v", err)
		return
	}
	log.Infof("Parsed part:%v,  %v", video.Url, title)
}

func parseUrls(urlsText string) *[]storage.VideoUrl {
	parts := strings.Split(urlsText, ",")
	var urls []storage.VideoUrl
	for _, p := range parts {
		firstPart := strings.Split(p, " or ")[0]
		mp4url := strings.ReplaceAll(strings.Split(p, " or ")[1], "\\", "")
		qualityPre := strings.Split(firstPart, "]")[0]
		m3u8url := strings.ReplaceAll(strings.Split(firstPart, "]")[1], "\\", "")
		quality := qualityPre[1:len(qualityPre)]
		urls = append(urls, storage.VideoUrl{quality, mp4url, m3u8url})
	}
	sortByQuality(urls)
	return &urls
}

func sortByQuality(urls []storage.VideoUrl) {
	sort.Slice(urls, func(i, j int) bool {
		iC := strings.ReplaceAll(urls[i].Quality, "p", "")
		iC = strings.ReplaceAll(iC, " Ultra", "0")
		iV, err := strconv.Atoi(iC)
		if err != nil {
			return false
		}
		jC := strings.ReplaceAll(urls[j].Quality, "p", "")
		jC = strings.ReplaceAll(jC, " Ultra", "0")
		jV, err := strconv.Atoi(jC)
		if err != nil {
			return false
		}
		return iV > jV
	})
}
