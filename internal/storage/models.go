package storage

import (
	"encoding/json"
	"go_rezka/internal/helpers"
)

type Genre struct {
	Id   uint
	Name string
}

type Video struct {
	Id          uint
	Name        string
	Name_orig   string
	Url         string
	Image_url   string
	Description string
	Rating      float64
	Video_urls  string
}

type VideoUrl struct {
	Quality string `json:"quality"`
	Mp4url  string `json:"mp4Url"`
	M3u8url string `json:"m3u8url"`
}

func (video Video) GetUrls() (urls []VideoUrl, err error) {
	err = json.Unmarshal([]byte(video.Video_urls), &urls)
	if err != nil {
		return nil, err
	}
	helpers.ReverseAny(urls)
	return urls, nil
}
