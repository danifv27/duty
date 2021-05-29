package main

import (
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/danifv27/duty/internal/duty"

	log "github.com/gomicro/ledger"
)

var (
	conf *duty.File
	// proxies map[string]*httputil.ReverseProxy
)

func configure() {
	c, err := duty.ParseFromFile()
	if err != nil {
		log.Fatalf("Failed to read config file: %v", err.Error())
		os.Exit(1)
	}

	conf = c
	log.Debug("Config file parsed")
	// register with the prometheus collector
	prometheus.MustRegister(duty.TotalRequests)
	log.Debug("Prometheus handler registered")
	log.Debug("Configuration complete")
}

func main() {
	configure()

	promMux := http.NewServeMux()
	promMux.Handle("/prometheus", promhttp.Handler())

	go func() {
		log.Infof("Listening on %v:%v", "0.0.0.0", "9000")
		http.ListenAndServe(":9000", promMux)
	}()

	log.Infof("Listening on %v:%v", "0.0.0.0", "4567")
	http.ListenAndServe("0.0.0.0:4567", conf)
}
