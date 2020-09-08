package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"sort"
	"time"

	"github.com/apex/log"
	"github.com/apex/log/handlers/cli"
	"github.com/dustin/go-humanize"
	"github.com/jroimartin/gocui"
	"github.com/midgarco/movie_downloader/config"
	"github.com/midgarco/movie_downloader/log/gui"
	"github.com/midgarco/movie_downloader/rpc/moviedownloader"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

var (
	port       = flag.String("p", "8080", "port to run the agent on")
	endpoint   = flag.String("endpoint", "", "pmd server connection")
	configFile = flag.String("config", os.Getenv("HOME")+"/.pmd/agent.yaml", "The path to the config.yaml file")

	client moviedownloader.MovieDownloaderServiceClient
)

func init() {
	flag.Parse()
}

func main() {
	log.SetLevel(log.DebugLevel)
	log.SetHandler(cli.New(os.Stdout))

	// load the configuration
	viper.SetConfigName("agent")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(path.Dir(*configFile))
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found so create one
			if err := config.Create(*configFile); err != nil {
				log.WithError(err).Error("failed to create config file")
			}

			// ask for the server endpoint
			viper.Set("GRPC_ENDPOINT", config.GetGRPCEndpoint(""))
		} else {
			log.WithError(err).Fatal("could not read in the config file")
		}
	}

	if *endpoint == "" {
		if viper.GetString("GRPC_ENDPOINT") == "" {
			viper.Set("GRPC_ENDPOINT", config.GetGRPCEndpoint(""))
		}
		*endpoint = viper.GetString("GRPC_ENDPOINT")
	}

	if *endpoint == "" {
		log.Fatal("no GRPC endpoint configured")
	}

	viper.Set("GRPC_ENDPOINT", *endpoint)

	// update the configuration file
	if err := viper.WriteConfig(); err != nil {
		log.WithError(err).Error("failed to write config file")
	}

	var opts []grpc.DialOption
	var kacp = keepalive.ClientParameters{
		Time: time.Duration(10) * time.Second,
		// Timeout:             time.Second,
		// PermitWithoutStream: true,
	}
	opts = append(opts, grpc.WithInsecure(), grpc.WithKeepaliveParams(kacp))
	conn, err := grpc.Dial(*endpoint, opts...)
	if err != nil {
		log.WithFields(log.Fields{
			"endpoint": *endpoint,
		}).WithError(err).Fatal("failed to connect to the cap service client")
	}
	client = moviedownloader.NewMovieDownloaderServiceClient(conn)
	defer conn.Close()

	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer g.Close()

	g.SetManagerFunc(layout)

	// listen for web requests
	go func() {
		http.HandleFunc("/", handler)
		http.HandleFunc("/search", searchHandler)
		http.HandleFunc("/download", downloadHandler)

		log.Info("Agent listening on port :" + *port)
		if err := http.ListenAndServe(":"+*port, nil); err != nil {
			log.WithError(err).Fatal("failed to start agent service")
		}
	}()

	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		log.Fatal(err.Error())
	}

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Fatal(err.Error())
	}
}

func layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	if v, err := g.SetView("log", -1, 5, maxX, maxY); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Wrap = true
		v.Autoscroll = true

		log.SetHandler(gui.New(g, cli.New(v)))
		log.Info("hello world")
	}
	if v, err := g.SetView("downloads", 0, 0, maxX-1, 5); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Wrap = true
		v.Title = "Downloads"

		// display download content
		go func() {
			stream, err := client.Progress(context.Background(), &moviedownloader.ProgressRequest{})
			if err != nil {
				log.WithError(err).Fatal("error connecting to pmd service")
				return
			}
			log.Info("connected to progress stream")
			for {
				res, err := stream.Recv()
				if err == io.EOF {
					log.Warn("server connection lost")
					break
				}
				if err != nil {
					log.WithError(err).Error("failure receiving progress updates")
				}
				v.Clear()

				if res == nil || res.ActiveDownloads == nil {
					continue
				}

				keys := make([]int, 0, len(res.ActiveDownloads))
				for k := range res.ActiveDownloads {
					keys = append(keys, int(k))
				}
				sort.Ints(keys)

				// for idx, movie := range res.ActiveDownloads {
				for _, idx := range keys {
					movie := res.ActiveDownloads[int32(idx)]
					_, _ = v.Write([]byte(fmt.Sprintf("%d: (%d%%) %s/s %s/%s : %s\n",
						idx,
						movie.Progress,
						humanize.Bytes(uint64(movie.BytesPerSecond)),
						humanize.Bytes(uint64(movie.BytesCompleted)),
						humanize.Bytes(uint64(movie.Size)),
						movie.Filename,
					)))
				}
				g.Update(func(g *gocui.Gui) error { return nil })
			}
		}()
	}
	return nil
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

// SearchRequest ...
type SearchRequest struct {
	Query      string `json:"query"`
	ServerPath string `json:"server_path"`
}

// DownloadRequest ...
type DownloadRequest struct {
	Movie      moviedownloader.Movie `json:"movie"`
	ServerPath string                `json:"server_path"`
}

func handler(w http.ResponseWriter, r *http.Request) {
	log.Info("catch-all handler")
	fmt.Fprint(w, "PMD Version 1.0")
}

func searchHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "expected POST", 400)
		return
	}

	sr := SearchRequest{}
	err := json.NewDecoder(r.Body).Decode(&sr)
	if err != nil {
		http.Error(w, "could not decode request", 500)
		log.WithError(err).Error("decoding search request")
		return
	}

	log.Infof("searching for \"%s\"", sr.Query)

	// send the request to the server
	b := new(bytes.Buffer)
	err = json.NewEncoder(b).Encode(sr)
	if err != nil {
		http.Error(w, "could not encode json payload", 500)
		log.WithError(err).Error("error encoding payload")
		return
	}
	res, err := http.Post(fmt.Sprintf("%s/search", sr.ServerPath), "application/json; charset=utf-8", b)
	if err != nil {
		http.Error(w, "error response from server", 500)
		log.WithError(err).Error("server error")
		return
	}
	defer res.Body.Close()

	w.Header().Add("Content-type", "application/json")
	io.Copy(w, res.Body)
}

func downloadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "expected POST", 400)
		return
	}

	dr := DownloadRequest{}
	err := json.NewDecoder(r.Body).Decode(&dr)
	if err != nil {
		http.Error(w, "could not decode request", 500)
		log.WithError(err).Error("decoding download request")
		return
	}

	log.Infof("downloading \"%s\"", dr.Movie.Filename)

	// send the request to the server
	b := new(bytes.Buffer)
	err = json.NewEncoder(b).Encode(dr)
	if err != nil {
		http.Error(w, "could not encode json payload", 500)
		log.WithError(err).Error("error encoding payload")
		return
	}
	res, err := http.Post(fmt.Sprintf("%s/download", dr.ServerPath), "application/json; charset=utf-8", b)
	if err != nil {
		http.Error(w, "error response from server", 500)
		log.WithError(err).Error("server error")
		return
	}
	defer res.Body.Close()

	w.Header().Add("Content-type", "application/json")
	io.Copy(w, res.Body)
}
