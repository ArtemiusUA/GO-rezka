package storage

import (
	"encoding/json"
	"github.com/jmoiron/sqlx/types"
)

type Genre struct {
	Id   uint
	Type string
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
	Video_urls  types.JSONText
}

type Part struct {
	Id         uint
	Name       string
	Video_urls types.JSONText
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
	return urls, nil
}

func (part Part) GetUrls() (urls []VideoUrl, err error) {
	err = json.Unmarshal([]byte(part.Video_urls), &urls)
	if err != nil {
		return nil, err
	}
	return urls, nil
}
