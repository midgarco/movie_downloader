package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/midgarco/movie_downloader/server/movie"
)

var (
	port = flag.Int("p", 8080, "port to run the agent on")
)

func init() {
	flag.Parse()
}

// SearchRequest ...
type SearchRequest struct {
	Query      string `json:"query"`
	ServerPath string `json:"server_path"`
}

// DownloadRequest ...
type DownloadRequest struct {
	Movie      movie.Movie `json:"movie"`
	ServerPath string      `json:"server_path"`
}

func handler(w http.ResponseWriter, r *http.Request) {
	log.Infof("handler: %v\n", r)
	fmt.Fprint(w, "PMD Version 1.0")
}

func searchHandler(w http.ResponseWriter, r *http.Request) {
	log.Infof("searchHandler: %v\n", r)

	if r.Method != "POST" {
		http.Error(w, "expected POST", 400)
		return
	}

	sr := SearchRequest{}
	err := json.NewDecoder(r.Body).Decode(&sr)
	if err != nil {
		http.Error(w, "could not decode request", 400)
		log.Errorf("decoding search request: %v\n", err)
		return
	}

	// send the request to the server
	b := new(bytes.Buffer)
	err = json.NewEncoder(b).Encode(sr)
	if err != nil {
		http.Error(w, "could not encode json payload", 400)
		log.Errorf("encode payload: %v\n", err)
		return
	}
	res, err := http.Post(fmt.Sprintf("%s/search", sr.ServerPath), "application/json; charset=utf-8", b)
	if err != nil {
		http.Error(w, "error response from server", 400)
		log.Errorf("server response: %v\n", err)
		return
	}
	defer res.Body.Close()

	w.Header().Add("Content-type", "application/json")
	io.Copy(w, res.Body)
}

func downloadHandler(w http.ResponseWriter, r *http.Request) {
	log.Infof("downloadHandler: %v\n", r)

	if r.Method != "POST" {
		http.Error(w, "expected POST", 400)
		return
	}

	dr := DownloadRequest{}
	err := json.NewDecoder(r.Body).Decode(&dr)
	if err != nil {
		http.Error(w, "could not decode request", 400)
		log.Errorf("decoding download request: %v\n", err)
		return
	}

	// send the request to the server
	b := new(bytes.Buffer)
	err = json.NewEncoder(b).Encode(dr)
	if err != nil {
		http.Error(w, "could not encode json payload", 400)
		log.Errorf("encode payload: %v\n", err)
		return
	}
	res, err := http.Post(fmt.Sprintf("%s/download", dr.ServerPath), "application/json; charset=utf-8", b)
	if err != nil {
		http.Error(w, "error response from server", 400)
		log.Errorf("server response: %v\n", err)
		return
	}
	defer res.Body.Close()

	w.Header().Add("Content-type", "application/json")
	io.Copy(w, res.Body)
}

func main() {
	http.HandleFunc("/", handler)
	http.HandleFunc("/search", searchHandler)
	http.HandleFunc("/download", downloadHandler)

	log.Infof("Agent listening on port %d", *port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), nil))
}
