package movie

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/cavaliercoder/grab"
	humanize "github.com/dustin/go-humanize"
	"github.com/jeffdupont/movie_downloader/server/config"
)

const (
	// {url}/dl/auto/80/{file ID}{file extension}/{file name}{file extension}
	downloadURL string = "https://members.easynews.com/dl/auto/80/%s%s/%s%[2]s"
)

// Movie ...
type Movie struct {
	// One         string   `json:"1"`
	// One1        string   `json:"11"`
	// One3        string   `json:"13"`
	// One9        string   `json:"19"`
	ID          string   `json:"0"`
	Filename    string   `json:"10"`
	VideoCodec  string   `json:"12,omitempty"`
	Runtime     string   `json:"14,omitempty"`
	BPS         int      `json:"15,omitempty"`
	SampleRate  int      `json:"16,omitempty"`
	FPS         float64  `json:"17,omitempty"`
	AudioCodec  string   `json:"18,omitempty"`
	Extension   string   `json:"2"`
	ExpireDate  int      `json:"20,omitempty"`
	Resolution  string   `json:"3,omitempty"`
	Three5      string   `json:"35,omitempty"`
	Size        string   `json:"4,omitempty"`
	PostDate    string   `json:"5,omitempty"`
	Subject     string   `json:"6,omitempty"`
	Poster      string   `json:"7,omitempty"`
	Eight       string   `json:"8,omitempty"`
	Group       string   `json:"9,omitempty"`
	Alangs      []string `json:"alangs,omitempty"`
	Expires     string   `json:"expires,omitempty"`
	FallbackURL string   `json:"fallbackURL,omitempty"`
	Fullres     string   `json:"fullres,omitempty"`
	Height      string   `json:"height,omitempty"`
	Nfo         string   `json:"nfo,omitempty"`
	Passwd      bool     `json:"passwd,omitempty"`
	PrimaryURL  string   `json:"primaryURL,omitempty"`
	RawSize     int      `json:"rawSize,omitempty"`
	Sb          int      `json:"sb,omitempty"`
	Sc          string   `json:"sc,omitempty"`
	Slangs      []string `json:"slangs,omitempty"`
	Theight     int      `json:"theight,omitempty"`
	Ts          int      `json:"ts,omitempty"`
	Twidth      int      `json:"twidth,omitempty"`
	Type        string   `json:"type,omitempty"`
	Virus       bool     `json:"virus,omitempty"`
	Volume      bool     `json:"volume,omitempty"`
	Width       string   `json:"width,omitempty"`
}

// Download ...
func download(ctx context.Context, movie *Movie, path string) error {
	log.Infof("download request: %v", movie.ID)

	if movie.Virus {
		return fmt.Errorf("movie contained virus: %v", movie)
	}

	if movie.ID == "" || movie.Extension == "" || movie.Filename == "" {
		// http.Error(w, "malformed request", 400)
		return fmt.Errorf("malformed movie: %v", movie)
	}

	uri := fmt.Sprintf(downloadURL, movie.ID, movie.Extension, movie.Filename)

	request, err := grab.NewRequest(uri)
	if err != nil {
		return fmt.Errorf("grab request: %v", err)
	}

	cfg := ctx.Value(config.Configuration{}).(*config.Configuration)
	request.HTTPRequest.SetBasicAuth(cfg.Username, cfg.Password)
	request.Filename = path + movie.Filename + movie.Extension

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

	go func() {
		respch := client.DoAsync(request)
		resp := <-respch

		log.Infof("\nDownloading %s%s\n\n", movie.Filename, movie.Extension)

		// print progress until transfer is complete
		for !resp.IsComplete() {
			fmt.Printf("\033[1A%s/s %s / %s (%d%%)\033[K\n",
				humanize.Bytes(uint64(resp.AverageBytesPerSecond())), humanize.Bytes(resp.BytesTransferred()), humanize.Bytes(resp.Size), int(100*resp.Progress()))
			time.Sleep(200 * time.Millisecond)
		}

		// clear progress line
		fmt.Printf("\033[1A\033[K")

		// check for errors
		if resp.Error != nil {
			fmt.Fprintf(os.Stderr, "Error downloading %s: %v\n", uri, resp.Error)
			return
		}

		log.Infof("Successfully downloaded to %s\n", resp.Filename)
	}()

	return nil
}

type downloadRequest struct {
	Movie Movie `json:"movie"`
}

// Handler ...
func Handler(path string, w http.ResponseWriter, r *http.Request) {
	log.Infof("download.Handler: %v\n", r)

	if r.Body == nil {
		http.Error(w, "no data received", 400)
		return
	}

	dr := &downloadRequest{}
	err := json.NewDecoder(r.Body).Decode(&dr)
	if err != nil {
		http.Error(w, "unable to decode message", 400)
		log.Errorf("decoding json: %v\n", err)
		return
	}

	err = download(r.Context(), &dr.Movie, path)
	if err != nil {
		http.Error(w, "could not download file", 400)
		log.Errorf("download error: %v\n", err)
		return
	}
}
