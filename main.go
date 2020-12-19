package main

import (
	"context"
	"encoding/json"
	"flag"
	"io"
	"os"
	"path"
	"time"

	"github.com/apex/log"
	"github.com/leaanthony/mewn"
	"github.com/midgarco/movie_downloader/config"
	"github.com/midgarco/movie_downloader/rpc/moviedownloader"
	"github.com/spf13/viper"
	"github.com/wailsapp/wails"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

var (
	configFile = flag.String("config", os.Getenv("HOME")+"/.pmd/agent.yaml", "The path to the config.yaml file")
)

func init() {
	flag.Parse()
}

func main() {

	js := mewn.String("./frontend/dist/app.js")
	css := mewn.String("./frontend/dist/app.css")

	agent, err := NewAgent()
	if err != nil {
		agent.log.Fatalf("Failed to start Agent: %v", err)
	}

	app := wails.CreateApp(&wails.AppConfig{
		Width:     1024,
		Height:    768,
		Title:     "PMD Agent",
		JS:        js,
		CSS:       css,
		Colour:    "#FFFFFF",
		Resizable: true,
	})
	app.Bind(agent)
	app.Run()
}

type Agent struct {
	endpoint  string
	runtime   *wails.Runtime
	log       *wails.CustomLogger
	conn      *grpc.ClientConn
	downloads map[int32]*moviedownloader.Progress
}

func NewAgent() (*Agent, error) {
	agent := &Agent{}
	return agent, nil
}

func (a *Agent) WailsInit(runtime *wails.Runtime) error {
	// connect the wails runtime
	a.runtime = runtime
	a.log = runtime.Log.New("Agent")

	// load the configuration
	viper.SetConfigName("agent")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(path.Dir(*configFile))
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found so create one
			if err := config.Create(*configFile); err != nil {
				a.log.Errorf("Failed to create config file: %v", err)
			}
		} else {
			a.log.Fatalf("Could not read in the config file: %v", err)
		}
	}

	// update the configuration file
	if err := viper.WriteConfig(); err != nil {
		a.log.Errorf("Failed to write config file: %v", err)
	}

	// set the endpoint
	a.endpoint = viper.GetString("GRPC_ENDPOINT")

	// build out the GRPC connection
	if err := a.createConnection(); err != nil {
		return err
	}

	return a.Progress()
}

func (a *Agent) GetEndpoint() string {
	return viper.GetString("GRPC_ENDPOINT")
}

func (a *Agent) SaveEndpoint(endpoint string) error {

	viper.Set("GRPC_ENDPOINT", endpoint)

	// update the configuration file
	if err := viper.WriteConfig(); err != nil {
		a.log.Errorf("Failed to write config file: %v", err)
		return err
	}

	// set the endpoint
	a.endpoint = viper.GetString("GRPC_ENDPOINT")

	// build out the GRPC connection
	if err := a.createConnection(); err != nil {
		return err
	}

	return a.Progress()
}

func (a *Agent) createConnection() error {
	var opts []grpc.DialOption
	var kacp = keepalive.ClientParameters{
		Time: time.Duration(10) * time.Second,
		// Timeout:             time.Second,
		// PermitWithoutStream: true,
	}
	opts = append(opts, grpc.WithInsecure(), grpc.WithKeepaliveParams(kacp))
	conn, err := grpc.Dial(a.endpoint, opts...)
	if err != nil {
		a.log.Error(err.Error())
		return err
	}
	a.conn = conn
	return nil
}

func (a *Agent) Search(query string) (*moviedownloader.SearchResponse, error) {
	client := moviedownloader.NewMovieDownloaderServiceClient(a.conn)

	req := &moviedownloader.SearchRequest{Query: query}
	results, err := client.Search(context.Background(), req)
	if err != nil {
		return nil, err
	}
	return results, nil
}

func (a *Agent) Download(selected string) error {
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

func (a *Agent) Progress() error {
	a.log.Info("Start monitor active downloads")

	client := moviedownloader.NewMovieDownloaderServiceClient(a.conn)

	// display download content
	go func(client moviedownloader.MovieDownloaderServiceClient) {
		stream, err := client.Progress(context.Background(), &moviedownloader.ProgressRequest{})
		if err != nil {
			a.log.Errorf("Error connecting to pmd service: %v", err)
			return
		}
		log.Infof("Connected to progress stream %s", viper.GetString("GRPC_ENDPOINT"))
		for {
			res, err := stream.Recv()
			if err == io.EOF {
				a.log.Warn("Server connection lost")
				break
			}
			if err != nil {
				a.log.Errorf("Failure receiving progress updates: %v", err)
				continue
			}

			if res == nil {
				continue
			}

			a.downloads = res.ActiveDownloads
			a.runtime.Events.Emit("progress", a.downloads)
		}
	}(client)

	return nil
}

func (a *Agent) Complete(id int32) error {
	client := moviedownloader.NewMovieDownloaderServiceClient(a.conn)

	req := &moviedownloader.CompletedRequest{CompletedId: id}
	_, err := client.Completed(context.Background(), req)
	if err != nil {
		return err
	}
	return nil
}
