package duty

import (
	"context"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// var TotalRequests = prometheus.NewCounterVec(
// 	prometheus.CounterOpts{
// 		Name: "http_server_requests_seconds_count",
// 		Help: "Total number of requests received.",
// 	},
// 	[]string{"method", "status", "uri"},
// )

var ServiceLatency = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "http_server_requests_seconds",
		Help:    "Sum of the the duration of every request.",
		Buckets: []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 7.5, 10},
	},
	[]string{"method", "status", "uri"},
)

func ServeMetrics(ctx context.Context) error {
	var err error

	// register with the prometheus collector
	// prometheus.MustRegister(TotalRequests)
	prometheus.MustRegister(ServiceLatency)
	promMux := http.NewServeMux()
	promMux.Handle("/prometheus", promhttp.Handler())
	promSrv := &http.Server{
		Addr:    ":9000",
		Handler: promMux,
	}

	go func() {
		log.Infof("Listening on %v", promSrv.Addr)
		if err = promSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %+s", err)
		}
	}()

	<-ctx.Done()

	log.Info("Metrics server stopped")
	ctxShutDown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		cancel()
	}()

	if err = promSrv.Shutdown(ctxShutDown); err != nil {
		log.Fatalf("metrics server Shutdown Failed:%+s", err)
	}
	log.Info("Metrics server shutdown")

	if err == http.ErrServerClosed {
		err = nil
	}

	return err
}
