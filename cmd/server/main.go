package main

import (
	"context"
	"flag"
	"net"
	"net/http"
	"net/http/pprof"
	"os"
	"path"
	"strings"

	"github.com/apex/log"
	"github.com/apex/log/handlers/cli"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/midgarco/movie_downloader/rpc/moviedownloader"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/encoding/protojson"
)

var (
	Version = "unset"
	Build   = "unset"

	configFile   = flag.String("config", os.Getenv("HOME")+"/.pmd/config.yaml", "The path to the config.yaml file")
	port         = flag.String("p", "4050", "The server REST port")
	grpcPort     = flag.String("grpc", "4051", "The server GRPC port")
	downloadPath = flag.String("d", os.Getenv("HOME")+"/Movies/", "The directory to save downloads")
)

func init() {
	flag.Parse()
}

func main() {
	log.SetLevel(log.DebugLevel)
	log.SetHandler(cli.New(os.Stdout))

	log.Log = log.WithFields(log.Fields{
		"version": Version,
		"build":   Build,
	})

	opts := &Options{
		DownloadPath: *downloadPath,
	}

	log.WithField("filename", *configFile).Info("loading config file")
	if err := srv.LoadConfig(*configFile, opts); err != nil {
		log.WithFields(log.Fields{
			"filename": *configFile,
		}).WithError(err).Fatal("failed loading config file")
	}

	log.WithFields(log.Fields{
		"rest_port":     *port,
		"grpc_port":     *grpcPort,
		"download_path": *downloadPath,
	}).Info("successfully loaded configuration")

	// start the REST proxy endpoints
	go func() {
		ctx := context.Background()
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		mux := runtime.NewServeMux(
			runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.HTTPBodyMarshaler{
				Marshaler: &runtime.JSONPb{
					MarshalOptions: protojson.MarshalOptions{
						UseProtoNames:   true,
						EmitUnpopulated: true,
					},
					UnmarshalOptions: protojson.UnmarshalOptions{
						DiscardUnknown: true,
					},
				},
			}),
		)
		opts := []grpc.DialOption{grpc.WithInsecure()}
		err := moviedownloader.RegisterMovieDownloaderServiceHandlerFromEndpoint(ctx, mux, ":"+*grpcPort, opts)
		if err != nil {
			log.WithError(err).Error("Failed to register handlers")
			return
		}

		httpmux := http.NewServeMux()

		// add support for pprof debugging and profiling
		httpmux.Handle("/debug/pprof/", http.HandlerFunc(pprof.Index))

		// add support for OpenAPI documentation
		httpmux.HandleFunc("/swagger/", func(w http.ResponseWriter, r *http.Request) {
			if !strings.HasSuffix(r.URL.Path, ".swagger.json") {
				log.Errorf("Not Found: %s", r.URL.Path)
				http.NotFound(w, r)
				return
			}

			log.Infof("Serving %s", r.URL.Path)
			p := strings.TrimPrefix(r.URL.Path, "/swagger/")
			p = path.Join("rpc", p)
			http.ServeFile(w, r, p)
		})
		httpmux.Handle("/", mux)

		log.Info("REST server listening on port :" + *port)
		if err := http.ListenAndServe(":"+*port, httpmux); err != nil {
			log.WithError(err).Error("Failed to start REST service")
			return
		}
	}()

	// start the GRPC endpoints
	lis, err := net.Listen("tcp", ":"+*grpcPort)
	if err != nil {
		log.WithError(err).Error("Failed to listen")
		return
	}

	s := grpc.NewServer()
	moviedownloader.RegisterMovieDownloaderServiceServer(s, srv)

	log.Info("GRPC server listening on port :" + *grpcPort)
	if err := s.Serve(lis); err != nil {
		log.WithError(err).Error("Failed to start GRPC service")
		return
	}
}
