package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/apex/log"
	"github.com/cavaliercoder/grab"
	"github.com/midgarco/movie_downloader/config"
	"github.com/midgarco/movie_downloader/cookiejar"
	"github.com/midgarco/movie_downloader/movie"
	"github.com/midgarco/movie_downloader/rpc/moviedownloader"
	"github.com/midgarco/movie_downloader/search"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// server represents the exporter
type server struct {
	Version string
	Build   string

	downloadPath        string
	searchUrlTemplate   string
	downloadUrlTemplate string
	cfg                 *config.Configuration

	activeDownloads map[string]*Download
}

type Download struct {
	BytesPerSecond int32
	BytesCompleted int32
	Size           int32
	Progress       int32
	Filename       string
	Details        *movie.Movie
	Error          string
}

type Options struct {
	DownloadPath string
}

var srv *server = &server{
	searchUrlTemplate:   "https://members.easynews.com/2.0/search/solr-search/?fly=2&gps=%s&pby=100&pno=1&s1=dtime&s1d=-&s2=nrfile&s2d=-&s3=dsize&s3d=-&sS=0&d1t=&d2t=&b1t=&b2t=&px1t=&px2t=&fps1t=&fps2t=&bps1t=&bps2t=&hz1t=&hz2t=&rn1t=&rn2t=&fty[]=VIDEO&u=1&sc=1&st=adv&safeO=0&sb=1",
	downloadUrlTemplate: "https://members.easynews.com/dl/auto/80/%s%s/%s%[2]s",
	activeDownloads:     map[string]*Download{},
}

// LoadConfig loads the configuration file into the server. If the files
// doesn't exist, it will prompt the user for the necessary credentials
// to create the file
func (s *server) LoadConfig(filename string, opts *Options) error {
	cfg, err := config.LoadConfig(filename)
	if err != nil {
		return err
	}

	s.cfg = cfg
	s.downloadPath = opts.DownloadPath

	return nil
}

// Search ...
func (s *server) Search(ctx context.Context, req *moviedownloader.SearchRequest) (*moviedownloader.SearchResponse, error) {
	log.Info("search: " + req.Query)

	uri := fmt.Sprintf(s.searchUrlTemplate, url.QueryEscape(req.Query))

	resp := &moviedownloader.SearchResponse{}

	defer func(resp *moviedownloader.SearchResponse) {
		log.WithFields(log.Fields{
			// "response": resp,
			"found": len(resp.Results.Movies),
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
		return nil, err
	}

	// set the basic auth
	request.SetBasicAuth(s.cfg.Username, s.cfg.Password)

	// save the authentication cookie
	cookies := &cookiejar.CookieJar{}
	cookies.Jar = make(map[string][]*http.Cookie)
	client.Jar = cookies

	// get the response from the client
	res, err := client.Do(request)
	if err != nil {
		log.WithError(err).Error("search request failed")
		return nil, err
	}
	defer res.Body.Close()

	results := &search.Results{}
	if err := json.NewDecoder(res.Body).Decode(results); err != nil {
		log.WithError(err).Error("decoding search response")
		return nil, err
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
		return nil, err
	}
	log.WithField("id", mv.ID).Info("download request")

	if mv.Virus {
		log.Error("attempted movie contains a virus")
		return nil, errors.New("movie contains virus")
	}

	if mv.ID == "" || mv.Extension == "" || mv.Filename == "" {
		log.WithField("movie", fmt.Sprintf("%#v", mv)).Error("malformed movie data")
		return nil, errors.New("malformed movie")
	}

	uri := fmt.Sprintf(s.downloadUrlTemplate, mv.ID, mv.Extension, mv.Filename)

	request, err := grab.NewRequest(".", uri)
	if err != nil {
		log.WithError(err).Error("creating grab request")
		return nil, errors.New("failed grab request")
	}

	request.HTTPRequest.SetBasicAuth(s.cfg.Username, s.cfg.Password)
	request.Filename = s.downloadPath + mv.Filename + mv.Extension

	// setup the net transport for tls
	var tran = &http.Transport{
		Dial: (&net.Dialer{
			Timeout: 5 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 5 * time.Second,
	}

	// establish the client for connection
	var httpClient = &http.Client{
		// Timeout:   10 * time.Second,
		Transport: tran,
	}

	client := grab.Client{
		HTTPClient: httpClient,
		UserAgent:  "grab",
	}

	dl := &Download{
		Filename: mv.Filename + mv.Extension,
		Details:  mv,
	}
	s.activeDownloads[mv.ID] = dl

	go func(mv *movie.Movie, stats *Download) {
		resp := client.Do(request)

		log.Info("downloading: " + mv.Filename + mv.Extension)

		// print progress until transfer is complete
		for !resp.IsComplete() {
			stats.BytesCompleted = int32(resp.BytesComplete())
			stats.BytesPerSecond = int32(resp.BytesPerSecond())
			stats.Size = int32(resp.Size)
			stats.Progress = int32(100 * resp.Progress())

			time.Sleep(200 * time.Millisecond)
		}

		// check for errors
		if resp.Err() != nil {
			log.WithError(resp.Err()).Error("download failed")
			stats.Error = resp.Err().Error()
			return
		}

		log.Info("successfully downloaded: " + resp.Filename)

		delete(s.activeDownloads, mv.ID)
	}(mv, dl)

	return &moviedownloader.Empty{}, nil
}

// Progress ...
func (s *server) Progress(req *moviedownloader.ProgressRequest, stream moviedownloader.MovieDownloaderService_ProgressServer) error {
	log.Info("starting progress stream")

	go func() {
		for {
			downloads := map[string]*moviedownloader.Progress{}
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