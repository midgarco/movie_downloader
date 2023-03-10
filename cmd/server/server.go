package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"sync"
	"time"

	"github.com/apex/log"
	"github.com/cavaliergopher/grab/v3"
	"github.com/midgarco/movie_downloader/config"
	"github.com/midgarco/movie_downloader/cookiejar"
	"github.com/midgarco/movie_downloader/movie"
	moviedownloader "github.com/midgarco/movie_downloader/rpc/api/v1"
	"github.com/midgarco/movie_downloader/search"
	"github.com/spf13/viper"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// server represents the exporter
type server struct {
	Version string
	Build   string

	downloadPath        string
	mediaPath           string
	searchUrlTemplate   string
	downloadUrlTemplate string

	mu                 sync.Mutex
	activeDownloads    map[int32]*Download
	completedDownloads map[int32]*Download
	downloadCount      int32
}

type Options struct{}

type Download struct {
	index int32

	BytesPerSecond int64
	BytesCompleted int64
	Size           int64
	Progress       int64
	Filename       string
	Details        *movie.Movie
	Error          string
}

var srv *server = &server{
	searchUrlTemplate:   "https://members.easynews.com/2.0/search/solr-search/?fly=2&gps=%s&pby=100&pno=1&s1=dtime&s1d=-&s2=nrfile&s2d=-&s3=dsize&s3d=-&sS=0&d1t=&d2t=&b1t=&b2t=&px1t=&px2t=&fps1t=&fps2t=&bps1t=&bps2t=&hz1t=&hz2t=&rn1t=&rn2t=&fty[]=VIDEO&u=1&sc=1&st=adv&safeO=0&sb=1",
	downloadUrlTemplate: "https://members.easynews.com/dl/auto/80/%s%s/%s%[2]s",
	activeDownloads:     map[int32]*Download{},
	completedDownloads:  map[int32]*Download{},
	downloadCount:       0,
}

// LoadConfig loads the configuration file into the server. If the files
// doesn't exist, it will prompt the user for the necessary credentials
// to create the file
func (s *server) LoadConfig(opts *Options) error {

	// load the configuration
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(path.Dir(*configFile))
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Info("creating config.yaml")

			// Config file not found so create one
			if err := config.Create(*configFile); err != nil {
				log.WithError(err).Error("failed to create config file")
			}

			// ask for the service credentials
			username, password := config.GetCredentials()
			viper.Set("USERNAME", username)
			viper.Set("PASSWORD", password)

			// ask for the download and media paths
			viper.Set("DOWNLOAD_PATH", *downloadPath)
			viper.Set("MEDIA_PATH", *mediaPath)
		} else {
			log.WithError(err).Fatal("could not read in the config file")
		}
	}

	if *downloadPath == "" {
		if viper.GetString("DOWNLOAD_PATH") == "" {
			viper.Set("DOWNLOAD_PATH", config.GetDownloadPath(""))
		}
		*downloadPath = viper.GetString("DOWNLOAD_PATH")
		viper.Set("DOWNLOAD_PATH", *downloadPath)
	}

	if *mediaPath == "" {
		if viper.GetString("MEDIA_PATH") == "" {
			viper.Set("MEDIA_PATH", config.GetMediaPath(""))
		}
		*mediaPath = viper.GetString("MEDIA_PATH")
		viper.Set("MEDIA_PATH", *mediaPath)
	}

	// update the configuration file
	if err := viper.WriteConfig(); err != nil {
		log.WithError(err).Error("failed to write config file")
	}

	s.downloadPath = viper.GetString("DOWNLOAD_PATH")
	s.mediaPath = viper.GetString("MEDIA_PATH")

	return nil
}

// Search ...
func (s *server) Search(ctx context.Context, req *moviedownloader.SearchRequest) (*moviedownloader.SearchResponse, error) {
	log.Info("search: " + req.Query)

	uri := fmt.Sprintf(s.searchUrlTemplate, url.QueryEscape(req.Query))

	resp := &moviedownloader.SearchResponse{}

	defer func(resp *moviedownloader.SearchResponse) {
		var movieCount int
		if resp.Results != nil {
			movieCount = len(resp.Results.Movies)
		}
		log.WithFields(log.Fields{
			// "response": resp,
			"found": movieCount,
		}).Info("search response")
	}(resp)

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
		log.WithError(err).Error("failed creating request")
		st := status.New(codes.Internal, "failed creating request")
		return nil, st.Err()
	}

	// set the basic auth
	request.SetBasicAuth(viper.GetString("USERNAME"), viper.GetString("PASSWORD"))

	// save the authentication cookie
	cookies := &cookiejar.CookieJar{}
	cookies.Jar = make(map[string][]*http.Cookie)
	client.Jar = cookies

	// get the response from the client
	res, err := client.Do(request)
	if err != nil {
		log.WithError(err).Error("search request failed")
		st := status.New(codes.Internal, "search request failed")
		return nil, st.Err()
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		log.Error(res.Status)
		st := status.New(codes.Code(res.StatusCode), res.Status)
		return nil, st.Err()
	}

	results := &search.Results{}
	if err := json.NewDecoder(res.Body).Decode(results); err != nil {
		log.WithError(err).Error("decoding search response")
		st := status.New(codes.Internal, "decoding search response")
		return nil, st.Err()
	}

	// format results for the response
	resp.Results = results.MapToProto()

	return resp, nil
}

// Download ...
func (s *server) Download(ctx context.Context, req *moviedownloader.DownloadRequest) (*moviedownloader.Empty, error) {
	mv, err := movie.MapFromProtoObject(req.Movie)
	if err != nil {
		log.WithError(err).Error("failed to map proto object")
		st := status.New(codes.Internal, "failed to map proto object")
		return nil, st.Err()
	}
	log.WithField("id", mv.ID).Info("download request")

	if mv.Virus {
		log.Error("attempted movie contains a virus")
		st := status.New(codes.FailedPrecondition, "movie contains virus")
		return nil, st.Err()
	}

	if mv.ID == "" || mv.Extension == "" || mv.Filename == "" {
		log.WithField("movie", fmt.Sprintf("%#v", mv)).Error("malformed movie data")
		st := status.New(codes.FailedPrecondition, "malformed movie data")
		return nil, st.Err()
	}

	uri := fmt.Sprintf(s.downloadUrlTemplate, mv.ID, mv.Extension, mv.Filename)
	log.Debug(uri)

	request, err := grab.NewRequest(".", uri)
	if err != nil {
		log.WithError(err).Error("creating grab request")
		st := status.New(codes.Internal, "failed grab request")
		return nil, st.Err()
	}

	request.HTTPRequest.SetBasicAuth(viper.GetString("USERNAME"), viper.GetString("PASSWORD"))
	request.Filename = filepath.Join(s.downloadPath, mv.Filename+mv.Extension)

	// setup the net transport for tls
	var tran = &http.Transport{
		Dial: (&net.Dialer{
			Timeout: 5 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 5 * time.Second,
	}

	// establish the client for connection
	var httpClient = &http.Client{
		Transport: tran,
	}

	client := grab.Client{
		HTTPClient: httpClient,
		UserAgent:  "grab",
	}

	s.downloadCount++
	dl := &Download{
		Filename: mv.Filename + mv.Extension,
		Details:  mv,
		index:    s.downloadCount,
	}
	s.activeDownloads[s.downloadCount] = dl

	go func(mv *movie.Movie, stats *Download) {
		resp := client.Do(request)

		log.Info("downloading: " + mv.Filename + mv.Extension)

		// start UI loop
		t := time.NewTicker(500 * time.Millisecond)
		defer t.Stop()

	Loop:
		for {
			select {
			case <-t.C:
				stats.BytesCompleted = resp.BytesComplete()
				stats.BytesPerSecond = int64(resp.BytesPerSecond())
				stats.Size = resp.Size()
				stats.Progress = int64(100 * resp.Progress())

			case <-resp.Done:
				// download is complete
				break Loop
			}
		}

		// check for errors
		if resp.Err() != nil {
			log.WithError(resp.Err()).Error("download failed")
			stats.Error = resp.Err().Error()
			return
		}

		if resp.IsComplete() {
			stats.Progress = 100
			stats.BytesCompleted = stats.Size
		}

		log.Info("successfully downloaded: " + resp.Filename)

		s.mu.Lock()
		delete(s.activeDownloads, stats.index)
		s.completedDownloads[stats.index] = stats
		s.mu.Unlock()
	}(mv, dl)

	return &moviedownloader.Empty{}, nil
}

// Progress ...
func (s *server) Progress(req *moviedownloader.ProgressRequest, stream moviedownloader.MovieDownloaderService_ProgressServer) error {
	log.Info("starting progress stream")

	go func() {
		for {
			downloads := map[int32]*moviedownloader.Progress{}
			// show the list of active downloads
			for id, dl := range s.activeDownloads {
				downloads[id] = &moviedownloader.Progress{
					BytesPerSecond: dl.BytesPerSecond,
					BytesCompleted: dl.BytesCompleted,
					Size:           dl.Size,
					Progress:       dl.Progress,
					Filename:       dl.Filename,
					Error:          dl.Error,
					Details:        dl.Details.MapToProto(),
				}
			}
			// show the list of completed downloads
			for id, dl := range s.completedDownloads {
				downloads[id] = &moviedownloader.Progress{
					BytesPerSecond: dl.BytesPerSecond,
					BytesCompleted: dl.BytesCompleted,
					Size:           dl.Size,
					Progress:       dl.Progress,
					Filename:       dl.Filename,
					Error:          dl.Error,
					Details:        dl.Details.MapToProto(),
				}
			}
			resp := &moviedownloader.ProgressResponse{
				ActiveDownloads: downloads,
			}

			if err := stream.Send(resp); err != nil {
				if status.Code(err) == codes.Unavailable {
					log.Info("stream closed")
					return
				}

				log.WithError(err).Error("failed to stream progress")
				return
			}

			time.Sleep(time.Second * 1)
		}
	}()

	<-stream.Context().Done()

	return nil
}

// Completed ...
func (s *server) Completed(ctx context.Context, req *moviedownloader.CompletedRequest) (*moviedownloader.CompletedResponse, error) {
	log.WithFields(log.Fields{
		"request": fmt.Sprintf("%#v", req),
	}).Info("completed request")

	resp := &moviedownloader.CompletedResponse{
		Completed: map[int32]*moviedownloader.Progress{},
	}
	defer func(resp *moviedownloader.CompletedResponse) {
		log.WithFields(log.Fields{
			"response": resp,
		}).Info("completed response")
	}(resp)

	// move the requested download to the media folder
	if req != nil && req.CompletedId > 0 {
		mv, ok := s.completedDownloads[req.CompletedId]
		if ok {
			filename := filepath.Join(s.downloadPath, mv.Filename)
			destfile := filepath.Join(s.mediaPath, mv.Filename)

			log.WithFields(log.Fields{
				"filename":    filename,
				"destination": destfile,
			}).Info("move the file")

			if err := os.Rename(filename, destfile); err != nil {
				log.WithError(err).Error("failed to move file")
			}

			s.mu.Lock()
			delete(s.completedDownloads, req.CompletedId)
			s.mu.Unlock()
		} else {
			log.Warn("could not find completed download")
		}
	}

	// list remaining completed items
	for id, dl := range s.completedDownloads {
		resp.Completed[id] = &moviedownloader.Progress{
			BytesPerSecond: dl.BytesPerSecond,
			BytesCompleted: dl.BytesCompleted,
			Size:           dl.Size,
			Progress:       dl.Progress,
			Filename:       dl.Filename,
			Error:          dl.Error,
			Details:        dl.Details.MapToProto(),
		}
	}

	return resp, nil
}
