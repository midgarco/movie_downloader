package movie

import (
	"github.com/midgarco/movie_downloader/rpc/moviedownloader"
)

// Movie ...
type Movie struct {
	// One         string   `json:"1"`
	// One1        string   `json:"11"`
	// One3        string   `json:"13"`
	// One9        string   `json:"19"`
	ID         string  `json:"0"`
	Filename   string  `json:"10"`
	VideoCodec string  `json:"12,omitempty"`
	Runtime    string  `json:"14,omitempty"`
	BPS        int     `json:"15,omitempty"`
	SampleRate int     `json:"16,omitempty"`
	FPS        float64 `json:"17,omitempty"`
	AudioCodec string  `json:"18,omitempty"`
	Extension  string  `json:"2"`
	// ExpireDate  int      `json:"20,omitempty"`
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

func (m Movie) MapToProto() *moviedownloader.Movie {
	return &moviedownloader.Movie{
		Id:             m.ID,
		Filename:       m.Filename,
		Extension:      m.Extension,
		Runtime:        m.Runtime,
		Resolution:     m.Resolution,
		FullResolution: m.Fullres,
		Size:           m.Size,
		Width:          m.Width,
		Height:         m.Height,
		Subject:        m.Subject,
		Group:          m.Group,
		PostDate:       m.PostDate,

		Codec:          m.VideoCodec,
		AudioCodec:     m.AudioCodec,
		AudioLanguages: m.Alangs,
		SubLanguages:   m.Slangs,
		Bps:            int32(m.BPS),
		SampleRate:     int32(m.SampleRate),
		Fps:            m.FPS,

		Virus: m.Virus,
		Type:  m.Type,
	}
}

func MapFromProtoObject(m *moviedownloader.Movie) (*Movie, error) {
	return &Movie{
		ID:          m.Id,
		Filename:    m.Filename,
		VideoCodec:  m.Codec,
		Runtime:     m.Runtime,
		BPS:         int(m.Bps),
		SampleRate:  int(m.SampleRate),
		FPS:         m.Fps,
		AudioCodec:  m.AudioCodec,
		Extension:   m.Extension,
		Resolution:  m.Resolution,
		Size:        m.Size,
		PostDate:    m.PostDate,
		Subject:     m.Subject,
		Poster:      m.Poster,
		Group:       m.Group,
		Alangs:      m.AudioLanguages,
		FallbackURL: m.FallbackUrl,
		Fullres:     m.FullResolution,
		Height:      m.Height,
		PrimaryURL:  m.PrimaryUrl,
		RawSize:     int(m.RawSize),
		Slangs:      m.SubLanguages,
		Type:        m.Type,
		Virus:       m.Virus,
		Width:       m.Width,
	}, nil
}
