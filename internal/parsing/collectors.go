package parsing

import (
	"encoding/json"
	"fmt"
	"github.com/ArtemiusUA/GO-rezka/internal/storage"
	"github.com/gocolly/colly/v2"
	"github.com/jmoiron/sqlx/types"
	log "github.com/sirupsen/logrus"
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

const userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/78.0.3904.108 Safari/537.36"

// CreateBaseCollector is creating main collector to parse lists of details urls
func CreateBaseCollector() *colly.Collector {
	videoCollector := CreateVideoCollector()

	baseCollector := colly.NewCollector(
		colly.UserAgent(userAgent),
		colly.AllowedDomains("rezka.ag", "www.rezka.ag"),
		colly.CacheDir("./cache"),
		colly.URLFilters(
			regexp.MustCompile(`https://rezka\.ag/films/$`),
			regexp.MustCompile(`https://rezka\.ag/films/page/.+/$`),
			regexp.MustCompile(`https://rezka\.ag/films/.+/.+\.html`),
			regexp.MustCompile(`https://rezka\.ag/cartoons/$`),
			regexp.MustCompile(`https://rezka\.ag/cartoons/page/.+/$`),
			regexp.MustCompile(`https://rezka\.ag/cartoons/.+/.+\.html`),
			regexp.MustCompile(`https://rezka\.ag/series/$`),
			regexp.MustCompile(`https://rezka\.ag/series/page/.+/$`),
			regexp.MustCompile(`https://rezka\.ag/series/.+/.+\.html`),
			regexp.MustCompile(`https://rezka\.ag/animation/$`),
			regexp.MustCompile(`https://rezka\.ag/animation/page/.+/$`),
			regexp.MustCompile(`https://rezka\.ag/animation/.+/.+\.html`),
		),
	)

	baseCollector.OnRequest(func(r *colly.Request) {
		r.Headers.Set("Accept", "*/*")
	})

	baseCollector.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		if regexp.MustCompile(`https://rezka\.ag/(films|cartoons|series|animation)/.+/.+\.html`).Match([]byte(link)) {
			videoCollector.Visit(link)
		} else {
			baseCollector.Visit(link)
		}
	})

	return baseCollector
}

// CreateVideoCollector is creating a main collector for detail video pages
func CreateVideoCollector() *colly.Collector {
	videoCollector := colly.NewCollector(colly.UserAgent(userAgent))

	videoCollector.OnRequest(func(r *colly.Request) {
		r.Headers.Set("Accept", "*/*")
		log.Debugf("Visiting: %v", r.URL.String())
	})
	videoCollector.OnHTML(".b-post", func(e *colly.HTMLElement) {
		name := e.ChildText(".b-content__main .b-post__title [itemprop=name]")
		nameOrig := e.ChildText(".b-content__main [itemprop=alternativeHeadline]")
		url := e.ChildAttr("[itemprop=url]", "content")
		imageUrl := e.ChildAttr("[itemprop=image]", "src")
		genres_names := e.ChildTexts("[itemprop=genre]")
		var genres []storage.Genre
		videoType := "films"
		typeMatch := regexp.MustCompile(`\/(\w+)\/`).FindSubmatch([]byte(e.Request.URL.Path))
		if typeMatch != nil && len(typeMatch) > 0 {
			videoType = string(typeMatch[1])
		}
		for _, g := range genres_names {
			genres = append(genres, storage.Genre{Type: videoType, Name: g})
		}

		description := e.ChildText(".b-post__description_text")
		rating, _ := strconv.ParseFloat(e.ChildText(".b-post__info_rates.imdb span"), 64)

		// parsing main default streams from the page itself
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

		// async fetching and parsing all related resources,
		// e.g. translations, series, etc.
		waitGroup := sync.WaitGroup{}
		if videoType == "films" {
			e.ForEach("#translators-list li", func(i int, e *colly.HTMLElement) {
				waitGroup.Add(1)
				go parseFilmPart(e, url, video, &waitGroup)
			})
		} else if videoType == "cartoons" || videoType == "series" || videoType == "animation" {
			var defaultTranslator string
			defaultTranslatorMatch := regexp.MustCompile(`initCDNSeriesEvents\(\d+,\s(\d+),`).FindSubmatch(e.Response.Body)
			if defaultTranslatorMatch != nil && len(defaultTranslatorMatch) > 0 {
				defaultTranslator = string(defaultTranslatorMatch[1])
			}
			e.ForEach("#simple-episodes-tabs li", func(i int, e *colly.HTMLElement) {
				waitGroup.Add(1)
				go parseSeriesPart(e, url, defaultTranslator, video, &waitGroup)
			})
			// TODO: Add logic for extra translations for series.
			// If the translation exists for series, then seasons and episodes list
			// are returned from the endpoint as HTML code string that should be
			// decoded and parsed.
		}
		waitGroup.Wait()

	})
	videoCollector.OnError(func(r *colly.Response, err error) {
		log.Errorf("Error parsing: %v, %v", r.Request.URL, err)
	})

	return videoCollector
}

func parseFilmPart(e *colly.HTMLElement, url string, video storage.Video, group *sync.WaitGroup) {
	defer group.Done()
	id := e.Attr("data-id")
	title := e.Attr("title")
	translatorId := e.Attr("data-translator_id")
	isCamrip := e.Attr("data-camrip")
	isAds := e.Attr("data-ads")
	isDirector := e.Attr("data-director")
	partUrl := fmt.Sprint("https://rezka.ag/ajax/get_cdn_series/?t=",
		time.Now().UnixNano()/int64(time.Millisecond))

	action := "get_movie"
	data := urlPackage.Values{
		"id":            {id},
		"translator_id": {translatorId},
		"is_director":   {isDirector},
		"is_camrip":     {isCamrip},
		"is_ads":        {isAds},
		"action":        {action},
	}

	resp, err := http.PostForm(partUrl, data)
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
		Id:         0,
		Name:       title,
		Video_urls: types.JSONText(urlsJSONText),
		Season_id:  0,
		Episode_id: 0,
	}
	err = storage.SaveVideoPart(&video, &part)
	if err != nil {
		log.Error("Error: %v", err)
		return
	}
	log.Infof("Parsed part:%v,  %v", video.Url, title)
}

func parseSeriesPart(e *colly.HTMLElement, url string, translatorId string, video storage.Video, group *sync.WaitGroup) {
	defer group.Done()
	id := e.Attr("data-id")
	title := e.Text
	seasonId := e.Attr("data-season_id")
	episodeId := e.Attr("data-episode_id")
	partUrl := fmt.Sprint("https://rezka.ag/ajax/get_cdn_series/?t=",
		time.Now().UnixNano()/int64(time.Millisecond))
	name := fmt.Sprintf("%v: %v", seasonId, title)

	action := "get_stream"
	data := urlPackage.Values{
		"id":            {id},
		"translator_id": {translatorId},
		"season":        {seasonId},
		"episode":       {episodeId},
		"action":        {action},
	}

	resp, err := http.PostForm(partUrl, data)
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
	seasonIdInt, err := strconv.Atoi(seasonId)
	if err != nil {
		seasonIdInt = 0
	}
	episodeIdInt, err := strconv.Atoi(episodeId)
	if err != nil {
		episodeIdInt = 0
	}
	part := storage.Part{
		Id:         0,
		Name:       name,
		Video_urls: types.JSONText(urlsJSONText),
		Season_id:  uint(seasonIdInt),
		Episode_id: uint(episodeIdInt),
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
