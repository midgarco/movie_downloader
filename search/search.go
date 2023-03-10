package search

import (
	"github.com/midgarco/movie_downloader/movie"
	moviedownloader "github.com/midgarco/movie_downloader/rpc/api/v1"
)

// Results ...
type Results struct {
	BaseURL       string        `json:"baseURL"`
	ClassicThumbs string        `json:"classicThumbs"`
	Movies        []movie.Movie `json:"data"`
	DlFarm        string        `json:"dlFarm"`
	DlPort        string        `json:"dlPort"`
	DownURL       string        `json:"downURL"`
	// Fields        map[string][]string `json:"fields"`
	// Groups        map[string][]int    `json:"groups"`
	GsColumns []struct {
		Name string `json:"name"`
		Num  int    `json:"num"`
	} `json:"gsColumns"`
	HInfo             int    `json:"hInfo"`
	Hidden            int    `json:"hidden"`
	Hthm              int    `json:"hthm"`
	LargeThumb        string `json:"largeThumb"`
	LargeThumbSize    string `json:"largeThumbSize"`
	NumPages          int    `json:"numPages"`
	Page              int    `json:"page"`
	PerPage           string `json:"perPage"`
	Count             int    `json:"results"`
	Returned          int    `json:"returned"`
	SS                string `json:"sS"`
	St                string `json:"st"`
	Stemmed           string `json:"stemmed"`
	ThumbURL          string `json:"thumbURL"`
	UnfilteredResults int    `json:"unfilteredResults"`
}

func (r Results) MapToProto() *moviedownloader.SearchResults {
	movies := []*moviedownloader.Movie{}
	for _, movie := range r.Movies {
		movies = append(movies, movie.MapToProto())
	}

	return &moviedownloader.SearchResults{
		// BaseUrl:  r.BaseURL,
		// DownUrl:  r.DownURL,
		Movies:   movies,
		Page:     int32(r.Page),
		NumPages: int32(r.NumPages),
		PerPage:  r.PerPage,
		Count:    int32(r.Count),
		Returned: int32(r.Returned),
	}
}
