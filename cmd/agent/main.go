package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/apex/log"
	"github.com/dustin/go-humanize"
	"github.com/marcusolsson/tui-go"
	"github.com/marcusolsson/tui-go/wordwrap"
	"github.com/midgarco/movie_downloader/log/channel"
	"github.com/midgarco/movie_downloader/rpc/moviedownloader"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

var (
	port     = flag.String("p", "8080", "port to run the agent on")
	endpoint = flag.String("endpoint", "", "pmd server connection")
)

func init() {
	flag.Parse()
}

func main() {
	logChan := make(chan string)

	log.SetLevel(log.DebugLevel)
	log.SetHandler(channel.New(logChan))

	if *endpoint == "" {
		reader := bufio.NewReader(os.Stdin)

		fmt.Print("Server GRPC endpoint: ")
		str, err := reader.ReadString('\n')
		if err != nil {
			log.WithError(err).Fatal("could not read server endpoint")
		}
		*endpoint = strings.TrimSpace(str)
	}

	// log window
	logs := tui.NewVBox()

	logScroll := tui.NewScrollArea(logs)
	logBox := tui.NewVBox(logScroll)
	logBox.SetTitle("logs")
	// logBox.SetBorder(true)
	logBox.SetSizePolicy(tui.Expanding, tui.Maximum)

	// download window
	downloads := tui.NewList()

	dlBox := tui.NewVBox(downloads)
	dlBox.SetTitle("downloads")
	dlBox.SetBorder(true)
	dlBox.SetSizePolicy(tui.Expanding, tui.Maximum)

	root := tui.NewVBox(dlBox, logBox)

	ui, err := tui.New(root)
	if err != nil {
		panic(err)
	}
	ui.SetKeybinding("Esc", func() { ui.Quit() })

	// display log content
	go func() {
		for txt := range logChan {
			boxSize := logBox.Size().X
			if boxSize < 80 {
				boxSize = 80
			}

			logs.Append(tui.NewHBox(
				tui.NewPadder(1, 0, tui.NewLabel(wordwrap.WrapString(txt, boxSize))),
				tui.NewSpacer(),
			))
			ui.Update(func() {})
		}
	}()

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
		}).WithError(err).Error("failed to connect to the cap service client")
	}
	client := moviedownloader.NewMovieDownloaderServiceClient(conn)
	defer conn.Close()

	// display download content
	go func(client moviedownloader.MovieDownloaderServiceClient) {
		stream, err := client.Progress(context.Background(), &moviedownloader.ProgressRequest{})
		if err != nil {
			log.WithError(err).Error("error connecting to pmd service")
			return
		}
		for {
			res, err := stream.Recv()
			if err == io.EOF {
				log.Warn("server connection lost")
				break
			}
			if err != nil {
				log.WithError(err).Error("failure receiving progress updates")
			}
			downloads.RemoveItems()

			for _, movie := range res.ActiveDownloads {
				downloads.AddItems(fmt.Sprintf(" (%d%%) %s/s %s/%s : %s",
					movie.Progress,
					humanize.Bytes(uint64(movie.BytesPerSecond)),
					humanize.Bytes(uint64(movie.BytesCompleted)),
					humanize.Bytes(uint64(movie.Size)),
					movie.Filename,
				))
			}
			ui.Update(func() {})
		}
	}(client)

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

	if err := ui.Run(); err != nil {
		panic(err)
	}
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
