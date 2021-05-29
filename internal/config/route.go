package config

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

const (
	ordinalRouteType  = "ordinal"
	staticRouteType   = "static"
	variableRouteType = "variable"
	verbRouteType     = "verb"
)

// Route represents a given endpoint and the kind of response it should return
type Route struct {
	Endpoint  string     `yaml:"endpoint"`
	Type      string     `yaml:"type"`
	Response  Response   `yaml:"response"`
	index     int        `yaml:"-"`
	Responses []Response `yaml:"responses"`
	Name      string     `yaml:"name"`
}

func (r *Route) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if req.Method == "OPTIONS" {
		r.handleCORS(w, req)
		return
	}

	switch strings.ToLower(r.Type) {
	case ordinalRouteType:
		r.handleOrdinalRoute(w, req)
		return

	case variableRouteType:
		r.handleVariableRoute(w, req)
		return

	case verbRouteType:
		r.handleVerbRoute(w, req)
		return

	default:
		r.handleDefaultRoute(w, req)
		return
	}
}

func (r *Route) handleCORS(w http.ResponseWriter, req *http.Request) {
	log.Info("responding with cors headers for options request")

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "*")
	w.Header().Set("Access-Control-Allow-Headers", "*, Authorization")
	w.Header().Set("Access-Control-Max-Age", "60")
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, post-check=0, pre-check=0")
	w.Header().Set("Vary", "Accept-Encoding")
	w.WriteHeader(http.StatusNoContent)

	return
}

func (r *Route) handleDefaultRoute(w http.ResponseWriter, req *http.Request) {
	var b []byte
	var err error

	if r.Response.Payload != "" {
		b, err = ioutil.ReadFile(r.Response.Payload)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("failed to read payload: %v", err.Error())))
			return
		}
	}

	w.WriteHeader(r.Response.Code)
	w.Write(b)
}

func (r *Route) handleOrdinalRoute(w http.ResponseWriter, req *http.Request) {
	var b []byte
	var err error

	i := r.index

	if r.Responses == nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("no payloads specified for ordinal endpoint"))
		return
	}

	if r.Responses[i].Payload != "" {
		b, err = ioutil.ReadFile(r.Responses[i].Payload)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("failed to read payload: %v", err.Error())))
			return
		}
	}

	w.WriteHeader(r.Responses[i].Code)
	w.Write(b)

	if i < len(r.Responses)-1 {
		r.index++
	}
}

func (r *Route) handleVariableRoute(w http.ResponseWriter, req *http.Request) {
	var b []byte
	var err error

	i := r.index

	if r.Responses == nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("no payloads specified for variable endpoint"))
		return
	}

	if r.Responses[i].Payload != "" {
		b, err = ioutil.ReadFile(r.Responses[i].Payload)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("failed to read payload: %v", err.Error())))
			return
		}
	}

	w.WriteHeader(r.Responses[i].Code)
	w.Write(b)
}

func (r *Route) handleVerbRoute(w http.ResponseWriter, req *http.Request) {
	for i := range r.Responses {
		if strings.ToUpper(r.Responses[i].Verb) == req.Method {
			var b []byte
			var err error

			if r.Responses[i].Payload != "" {
				b, err = ioutil.ReadFile(r.Responses[i].Payload)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte(fmt.Sprintf("failed to read payload: %v", err.Error())))
					return
				}
			}

			w.WriteHeader(r.Responses[i].Code)
			w.Write(b)
			return
		}
	}

	w.WriteHeader(http.StatusMethodNotAllowed)
	w.Write([]byte("method not defined in config"))
	return
}

// Reset returns the internal index of the route to 0
func (r *Route) Reset() {
	r.index = 0
}

// Set takes an id of the response desired, and sets the route to return the
// specified response if it exists. It will return an error if it is setting the
// response is not possible.
func (r *Route) Set(id string) error {
	if r.Type != variableRouteType {
		return fmt.Errorf("invalid route type")
	}

	for i, v := range r.Responses {
		if v.ID == id {
			r.index = i
			return nil
		}
	}

	return fmt.Errorf("ID not found")
}
