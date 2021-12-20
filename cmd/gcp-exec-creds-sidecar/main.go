package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/sirupsen/logrus"

	"github.com/romachalm/gcp-exec-creds-sidecar/pkg/gec"
)

var (
	verbose = flag.Bool("verbose", false, "debug traces")
	port    = flag.Int("port", 8080, "Port for HTTP server")
	iprange = flag.String("iprange", "127.0.0.1", "IPs to listen to")
)

func getHandler() (*chi.Mux, error) {
	// register gec as the handler
	gecs := &gec.GECServer{}
	r := chi.NewRouter()
	r.Mount("/", gec.Handler(gecs))
	return r, nil
}

func main() {

	flag.Parse()
	if os.Getenv("VERBOSE") == "1" || *verbose {
		logrus.SetLevel(logrus.DebugLevel)
		logrus.SetReportCaller(true)
		logrus.Debug("Let's talk !")
	}

	addr := fmt.Sprintf("%s:%d", *iprange, *port)
	logrus.Infof("Listening to %s", addr)

	handler, err := getHandler()
	if err != nil {
		logrus.WithError(err).Fatal("error initiating http server")
	}

	srv := &http.Server{
		Handler: handler,
		Addr:    addr,
	}

	logrus.Fatal(srv.ListenAndServe())
}
