package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path"
	"time"

	"github.com/apex/log"
	"github.com/apex/log/handlers/cli"
	"github.com/leaanthony/mewn"
	"github.com/midgarco/movie_downloader/config"
	"github.com/midgarco/movie_downloader/rpc/moviedownloader"
	"github.com/spf13/viper"
	"github.com/wailsapp/wails"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

var (
	endpoint   = flag.String("endpoint", "", "pmd server connection")
	configFile = flag.String("config", os.Getenv("HOME")+"/.pmd/agent.yaml", "The path to the config.yaml file")

	client moviedownloader.MovieDownloaderServiceClient

	downloads = make(map[int32]*moviedownloader.Progress)
)

func init() {
	flag.Parse()

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
}

func main() {

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

	// display download content
	go func() {
		stream, err := client.Progress(context.Background(), &moviedownloader.ProgressRequest{})
		if err != nil {
			log.WithError(err).Fatal("error connecting to pmd service")
			return
		}
		log.Infof("connected to progress stream %s", viper.GetString("GRPC_ENDPOINT"))
		for {
			res, err := stream.Recv()
			if err == io.EOF {
				log.Warn("server connection lost")
				break
			}
			if err != nil {
				log.WithError(err).Error("failure receiving progress updates")
				continue
			}

			if res == nil {
				log.WithField("result", fmt.Sprintf("%#v", res)).Warn("nil")
				continue
			}

			downloads = res.ActiveDownloads
		}
	}()

	js := mewn.String("./frontend/dist/app.js")
	css := mewn.String("./frontend/dist/app.css")

	app := wails.CreateApp(&wails.AppConfig{
		Width:     1024,
		Height:    768,
		Title:     "PMD Agent",
		JS:        js,
		CSS:       css,
		Colour:    "#FFFFFF",
		Resizable: true,
	})
	app.Bind(search)
	app.Bind(download)
	app.Bind(progress)
	app.Bind(complete)
	app.Run()
}

func search(query string) (*moviedownloader.SearchResponse, error) {
	req := &moviedownloader.SearchRequest{Query: query}
	results, err := client.Search(context.Background(), req)
	if err != nil {
		return nil, err
	}
	return results, nil
}

func download(selected string) error {
	movie := &moviedownloader.Movie{}
	if err := json.Unmarshal([]byte(selected), movie); err != nil {
		return err
	}

	req := &moviedownloader.DownloadRequest{Movie: movie}
	_, err := client.Download(context.Background(), req)
	if err != nil {
		return err
	}

	return nil
}

func progress() map[int32]*moviedownloader.Progress {
	return downloads
}

func complete(id int32) error {
	req := &moviedownloader.CompletedRequest{CompletedId: id}
	_, err := client.Completed(context.Background(), req)
	if err != nil {
		return err
	}
	return nil
}
