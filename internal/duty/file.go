package duty

import (
	"context"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/gomicro/ledger"
	"gopkg.in/yaml.v2"
)

const (
	defaultStatusEndpoint = "/duty/status"
	defaultResetEndpoint  = "/duty/reset"
	defaultSetEndpoint    = "/duty/set"
	defaultConfigFile     = "./duty.yaml"

	configFileEnv = "DUTY_CONFIG_FILE"
)

var (
	log  *ledger.Ledger
	conf *File
)

// File represents all the configurable options of Duty
type File struct {
	Routes    []Route           `yaml:"routes"`
	routesMap map[string]*Route `yaml:"-"`
	Status    string            `yaml:"status"`
	Reset     string            `yaml:"reset"`
	Set       string            `yaml:"set"`
}

func init() {
	log = ledger.New(os.Stdout, ledger.DebugLevel)
	rand.Seed(time.Now().UnixNano())
}

func configure() {
	c, err := ParseFromFile()
	if err != nil {
		log.Fatalf("Failed to read config file: %v", err.Error())
		os.Exit(1)
	}

	conf = c
	log.Debug("Config file parsed")
	// register with the prometheus collector
	prometheus.MustRegister(TotalRequests)
	prometheus.MustRegister(ServiceLatency)
	log.Debug("Prometheus handlers registered")
	log.Info("Configuration complete")
}

// ParseFromFile reads an Duty config file from the file specified in the
// environment or from the default file location if no environment is specified.
// A File with the populated values is returned and any errors encountered while
// trying to read the file.
func ParseFromFile() (*File, error) {
	configFile := os.Getenv(configFileEnv)

	if configFile == "" {
		configFile = defaultConfigFile
	}

	b, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err.Error())
	}

	var conf File
	err = yaml.Unmarshal(b, &conf)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal config file: %v", err.Error())
	}

	if conf.Status == "" {
		conf.Status = defaultStatusEndpoint
	}

	if conf.Reset == "" {
		conf.Reset = defaultResetEndpoint
	}

	if conf.Set == "" {
		conf.Set = defaultSetEndpoint
	}

	conf.routesMap = make(map[string]*Route)
	for i, r := range conf.Routes {
		conf.routesMap[r.Endpoint] = &conf.Routes[i]
	}

	return &conf, nil
}

func Serve(ctx context.Context) error {
	var err error

	configure()
	restSrv := &http.Server{
		Addr:    ":4567",
		Handler: conf,
	}
	go func() {
		log.Infof("Listening on %v", restSrv.Addr)
		if err = restSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen:%+s", err)
		}
	}()
	<-ctx.Done()

	log.Info("Server stopped")
	ctxShutDown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		cancel()
	}()

	if err = restSrv.Shutdown(ctxShutDown); err != nil {
		log.Fatalf("server Shutdown Failed:%+s", err)
	}
	log.Info("Server shutdown")

	if err == http.ErrServerClosed {
		err = nil
	}

	return err
}

func (f *File) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var status string
	var code int

	timer := prometheus.NewTimer(prometheus.ObserverFunc(func(v float64) {
		ServiceLatency.WithLabelValues(r.Method, status, r.URL.Path).Observe(v)
	}))
	defer func() {
		timer.ObserveDuration()
	}()

	defer func() {
		TotalRequests.WithLabelValues(r.Method, status, r.URL.Path).Inc()
	}()

	if r.URL.Path == f.Status {
		code = handleStatus(w, r)
		status = strconv.Itoa(code)
		return
	}

	if r.URL.Path == f.Reset {
		code = handleReset(w, r, f)
		status = strconv.Itoa(code)
		return
	}

	if r.URL.Path == f.Set {
		code = handleSet(w, r, f)
		status = strconv.Itoa(code)
		return
	}

	route, found := f.getRoute(r.URL)
	if !found {
		log.Errorf("route not found for url path: %v", r.URL)
		code = http.StatusNotFound
		w.WriteHeader(code)
		w.Write([]byte("path not found"))
		status = strconv.Itoa(code)
		return
	}

	code = route.ServeHTTP(w, r)
	status = strconv.Itoa(code)
}

func (f *File) getRoute(reqURL *url.URL) (*Route, bool) {
	r, found := f.routesMap[reqURL.Path]
	if !found {
		return nil, false
	}

	return r, true
}

func handleStatus(w http.ResponseWriter, req *http.Request) int {

	statusCode := http.StatusOK
	w.WriteHeader(statusCode)
	w.Write([]byte("duty is functioning"))

	return statusCode
}

func handleReset(w http.ResponseWriter, req *http.Request, f *File) int {
	log.Debug("resetting endpoints")

	statusCode := http.StatusOK
	for k := range f.routesMap {
		f.routesMap[k].Reset()
	}

	w.WriteHeader(statusCode)

	return statusCode
}

func handleSet(w http.ResponseWriter, req *http.Request, f *File) int {
	log.Debug("setting endpoint")

	name := req.URL.Query().Get("name")
	id := req.URL.Query().Get("id")
	statusCode := http.StatusBadRequest

	if name == "" || id == "" {
		w.WriteHeader(statusCode)
		w.Write([]byte("name and id are required query params"))
		return statusCode
	}

	for k := range f.routesMap {
		if f.routesMap[k].Name == name {
			err := f.routesMap[k].Set(id)
			if err != nil {
				statusCode = http.StatusInternalServerError
				w.WriteHeader(statusCode)
				w.Write([]byte(fmt.Sprintf("failed to set route: %v", err.Error())))
				return statusCode
			}
			statusCode = http.StatusOK
			w.WriteHeader(statusCode)
			return statusCode
		}
	}

	w.WriteHeader(statusCode)
	w.Write([]byte("no route found"))

	return statusCode
}
