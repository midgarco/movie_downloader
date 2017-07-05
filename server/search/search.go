package search

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/midgarco/movie_downloader/server/config"
	"github.com/midgarco/movie_downloader/server/cookiejar"
	"github.com/midgarco/movie_downloader/server/movie"
)

const (
	searchURL string = "https://members.easynews.com/2.0/search/solr-search/?fly=2&gps=%s&pby=100&pno=1&s1=dtime&s1d=-&s2=nrfile&s2d=-&s3=dsize&s3d=-&sS=0&d1t=&d2t=&b1t=&b2t=&px1t=&px2t=&fps1t=&fps2t=&bps1t=&bps2t=&hz1t=&hz2t=&rn1t=&rn2t=&fty[]=VIDEO&u=1&sc=1&st=adv&safeO=0&sb=1"
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

func search(ctx context.Context, query string) (*Results, error) {
	log.Infof("search query: %s\n", query)

	uri := fmt.Sprintf(searchURL, url.QueryEscape(query))

	// setup the net transport for tls
	var tran = &http.Transport{
		Dial: (&net.Dialer{
			Timeout: 5 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 5 * time.Second,
	}

	// establish the client for connection
	var client = &http.Client{
		Timeout:   10 * time.Second,
		Transport: tran,
	}

	// create the request
	request, err := http.NewRequest(http.MethodGet, uri, nil)
	if err != nil {
		log.Errorf("create search request: %v", err)
		return nil, err
	}

	cfg := ctx.Value(config.Configuration{}).(*config.Configuration)
	request.SetBasicAuth(cfg.Username, cfg.Password)

	// save the authentication cookie
	cookies := &cookiejar.CookieJar{}
	cookies.Jar = make(map[string][]*http.Cookie)
	client.Jar = cookies

	// get the response from the client
	res, err := client.Do(request)
	if err != nil {
		log.Errorf("search response: %v", err)
		return nil, err
	}
	defer res.Body.Close()

	results := Results{}
	if err := json.NewDecoder(res.Body).Decode(&results); err != nil {
		log.Errorf("decode search response: %v", err)
		return nil, err
	}

	return &results, nil
}

type searchRequest struct {
	Query string `json:"query"`
}

// Handler ...
func Handler(w http.ResponseWriter, r *http.Request) {
	log.Infof("search.Handler: %v", r)

	if r.Body == nil {
		http.Error(w, "no data received", 400)
		log.Error("no data received")
		return
	}

	sr := &searchRequest{}
	err := json.NewDecoder(r.Body).Decode(&sr)
	if err != nil {
		http.Error(w, "unable to decode message", 400)
		log.Errorf("decoding json: %v", err)
		return
	}

	results, err := search(r.Context(), sr.Query)
	if err != nil {
		http.Error(w, "could not perform search", 400)
		log.Errorf("search error: %v", err)
		return
	}

	w.Header().Add("Content-type", "application/json")
	json.NewEncoder(w).Encode(results)
}
