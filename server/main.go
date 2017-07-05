package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/midgarco/movie_downloader/server/config"
	"github.com/midgarco/movie_downloader/server/movie"
	"github.com/midgarco/movie_downloader/server/search"
)

var (
	configFile   = flag.String("config", os.Getenv("HOME")+"/.pmd/config.yaml", "The path to the config.yaml file")
	port         = flag.Int("p", 4050, "The server port")
	downloadPath = flag.String("d", os.Getenv("HOME")+"/Movies/", "The directory to save downloads")
	cfg          = config.Configuration{}
)

func init() {
	flag.Parse()
}

func handler(w http.ResponseWriter, r *http.Request) {
	log.Infof("handler: %v\n", r)

	if r.URL.Path != "/" {
		http.Error(w, "path not found", 404)
		return
	}
}

func downloadHandler(w http.ResponseWriter, r *http.Request) {
	movie.Handler(*downloadPath, w, r)
}

func addContext(next http.Handler) http.Handler {
	cfg, err := config.LoadConfig(*configFile)
	if err != nil {
		log.Errorf("loading config file %s: %v\n", configFile, err)
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.Method, "-", r.RequestURI)

		//Add data to context
		ctx := context.WithValue(r.Context(), config.Configuration{}, cfg)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/", handler)
	mux.HandleFunc("/search", search.Handler)
	mux.HandleFunc("/download", downloadHandler)

	contextedMux := addContext(mux)

	log.Printf("Server listening on port %d", *port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), contextedMux))
}
