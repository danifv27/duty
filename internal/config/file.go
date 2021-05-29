package config

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"

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
	log *ledger.Ledger
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
		return nil, fmt.Errorf("Failed to read config file: %v", err.Error())
	}

	var conf File
	err = yaml.Unmarshal(b, &conf)
	if err != nil {
		return nil, fmt.Errorf("Failed to unmarshal config file: %v", err.Error())
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

func (f *File) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == f.Status {
		handleStatus(w, r)
		return
	}

	if r.URL.Path == f.Reset {
		handleReset(w, r, f)
		return
	}

	if r.URL.Path == f.Set {
		handleSet(w, r, f)
		return
	}

	route, found := f.getRoute(r.URL)
	if !found {
		log.Errorf("route not found for url path: %v", r.URL)
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("path not found"))
		return
	}

	route.ServeHTTP(w, r)
}

func (f *File) getRoute(reqURL *url.URL) (*Route, bool) {
	r, found := f.routesMap[reqURL.Path]
	if !found {
		return nil, false
	}

	return r, true
}

func handleStatus(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("duty is functioning"))
}

func handleReset(w http.ResponseWriter, req *http.Request, f *File) {
	log.Debug("resetting endpoints")

	for k := range f.routesMap {
		f.routesMap[k].Reset()
	}

	w.WriteHeader(http.StatusOK)
	return
}

func handleSet(w http.ResponseWriter, req *http.Request, f *File) {
	log.Debug("setting endpoint")

	name := req.URL.Query().Get("name")
	id := req.URL.Query().Get("id")

	if name == "" || id == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("name and id are required query params"))
		return
	}

	for k := range f.routesMap {
		if f.routesMap[k].Name == name {
			err := f.routesMap[k].Set(id)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(fmt.Sprintf("failed to set route: %v", err.Error())))
				return
			}

			w.WriteHeader(http.StatusOK)
			return
		}
	}

	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte("no route found"))
	return
}
