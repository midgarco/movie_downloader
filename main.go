package main

import (
	"context"
	"embed"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path"
	"time"

	"github.com/apex/log"
	"github.com/midgarco/movie_downloader/config"
	"github.com/midgarco/movie_downloader/rpc/moviedownloader"
	"github.com/spf13/viper"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

var (
	configFile = flag.String("config", os.Getenv("HOME")+"/.pmd/agent.yaml", "The path to the config.yaml file")
)

//
func init() {
	flag.Parse()
}

//go:embed frontend/dist
var assets embed.FS

//
func main() {
	app := App{}
	err := wails.Run(&options.App{
		Title:  "PMD Agent",
		Width:  1024,
		Height: 768,
		Assets: assets,
		Bind: []interface{}{
			&app,
		},
		OnStartup:  app.startup,
		OnDomReady: app.domready,
	})
	if err != nil {
		log.WithError(err).Error("failed to start")
	}
}

//
type App struct {
	ctx context.Context

	endpoint  string
	conn      *grpc.ClientConn
	downloads map[int32]*moviedownloader.Progress
}

//
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// load the configuration
	viper.SetConfigName("agent")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(path.Dir(*configFile))
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found so create one
			if err := config.Create(*configFile); err != nil {
				runtime.LogError(a.ctx, fmt.Sprintf("Failed to create config file: %v", err))
			}
		} else {
			runtime.LogFatal(a.ctx, fmt.Sprintf("Could not read in the config file: %v", err))
		}
	}

	// update the configuration file
	if err := viper.WriteConfig(); err != nil {
		runtime.LogError(a.ctx, fmt.Sprintf("Failed to write config file: %v", err))
	}

	// set the endpoint
	a.endpoint = viper.GetString("GRPC_ENDPOINT")

	// build out the GRPC connection
	if err := a.createConnection(); err != nil {
		runtime.LogError(a.ctx, "failed to create grpc connection")

		os.Exit(1)
	}
}

//
func (a *App) GetEndpoint() string {
	return viper.GetString("GRPC_ENDPOINT")
}

//
func (a *App) SaveEndpoint(endpoint string) error {
	viper.Set("GRPC_ENDPOINT", endpoint)

	// update the configuration file
	if err := viper.WriteConfig(); err != nil {
		runtime.LogErrorf(a.ctx, "Failed to write config file: %v", err)
		return err
	}

	// set the endpoint
	a.endpoint = viper.GetString("GRPC_ENDPOINT")

	// build out the GRPC connection
	if err := a.createConnection(); err != nil {
		return err
	}

	a.domready(a.ctx)

	return nil
}

//
func (a *App) createConnection() error {
	var opts []grpc.DialOption
	var kacp = keepalive.ClientParameters{
		Time: time.Duration(10) * time.Second,
		// Timeout:             time.Second,
		// PermitWithoutStream: true,
	}
	opts = append(opts, grpc.WithInsecure(), grpc.WithKeepaliveParams(kacp))
	conn, err := grpc.Dial(a.endpoint, opts...)
	if err != nil {
		runtime.LogError(a.ctx, err.Error())
		return err
	}
	a.conn = conn
	return nil
}

//
func (a *App) Search(query string) (*moviedownloader.SearchResponse, error) {
	client := moviedownloader.NewMovieDownloaderServiceClient(a.conn)

	req := &moviedownloader.SearchRequest{Query: query}
	results, err := client.Search(context.Background(), req)
	if err != nil {
		return nil, err
	}
	return results, nil
}

//
func (a *App) Download(selected string) error {
	movie := &moviedownloader.Movie{}
	if err := json.Unmarshal([]byte(selected), movie); err != nil {
		return err
	}

	client := moviedownloader.NewMovieDownloaderServiceClient(a.conn)

	req := &moviedownloader.DownloadRequest{Movie: movie}
	_, err := client.Download(context.Background(), req)
	if err != nil {
		return err
	}

	return nil
}

//
func (a *App) domready(ctx context.Context) {
	runtime.LogInfo(a.ctx, "Start monitor active downloads")

	client := moviedownloader.NewMovieDownloaderServiceClient(a.conn)

	// display download content
	go func(client moviedownloader.MovieDownloaderServiceClient) {
		stream, err := client.Progress(ctx, &moviedownloader.ProgressRequest{})
		if err != nil {
			runtime.LogErrorf(a.ctx, "Error connecting to pmd service: %v", err)
			return
		}

		log.Infof("Connected to progress stream %s", viper.GetString("GRPC_ENDPOINT"))
		for {
			res, err := stream.Recv()
			if err == io.EOF {
				runtime.LogWarning(a.ctx, "Server connection lost")
				break
			}
			if err != nil {
				runtime.LogErrorf(a.ctx, "Failure receiving progress updates: %v", err)
				continue
			}

			if res == nil {
				continue
			}

			a.downloads = res.ActiveDownloads
			runtime.EventsEmit(ctx, "progress", a.downloads)
		}
	}(client)
}

//
func (a *App) Complete(id int32) error {
	client := moviedownloader.NewMovieDownloaderServiceClient(a.conn)

	req := &moviedownloader.CompletedRequest{CompletedId: id}
	_, err := client.Completed(context.Background(), req)
	if err != nil {
		return err
	}
	return nil
}
