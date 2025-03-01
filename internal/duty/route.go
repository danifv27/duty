package duty

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

const (
	ordinalRouteType  = "ordinal"
	staticRouteType   = "static"
	variableRouteType = "variable"
	verbRouteType     = "verb"
	rateRouteType     = "rate"
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

func (r *Route) ServeHTTP(w http.ResponseWriter, req *http.Request) int {

	if req.Method == "OPTIONS" {
		return r.handleCORS(w, req)
	}

	switch strings.ToLower(r.Type) {
	case ordinalRouteType:
		return r.handleOrdinalRoute(w, req)

	case variableRouteType:
		return r.handleVariableRoute(w, req)

	case verbRouteType:
		return r.handleVerbRoute(w, req)

	case rateRouteType:
		return r.handleRateRoute(w, req)

	default:
		return r.handleDefaultRoute(w, req)
	}
}

func (r *Route) handleCORS(w http.ResponseWriter, req *http.Request) int {

	statusCode := http.StatusNoContent

	log.Info("responding with cors headers for options request")

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "*")
	w.Header().Set("Access-Control-Allow-Headers", "*, Authorization")
	w.Header().Set("Access-Control-Max-Age", "60")
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, post-check=0, pre-check=0")
	w.Header().Set("Vary", "Accept-Encoding")
	w.WriteHeader(statusCode)

	return statusCode
}

func (r *Route) handleDefaultRoute(w http.ResponseWriter, req *http.Request) int {
	var b []byte
	var err error

	statusCode := http.StatusInternalServerError

	// log.Debugf("Response: %v", r.Response)
	if r.Response.Payload != "" {
		b, err = ioutil.ReadFile(r.Response.Payload)
		if err != nil {
			w.WriteHeader(statusCode)
			w.Write([]byte(fmt.Sprintf("failed to read payload: %v", err.Error())))
			return statusCode
		}
	}
	// log.Debug("Checking latency")
	if r.Response.Latency != "" {
		// log.Debug("Calculating latency")
		var lat time.Duration
		if strings.ToLower(r.Response.Latency) == "random" {
			lat = time.Duration(rand.Intn(9)) * time.Second // lat will be between 0 and 9
		} else if lat, err = time.ParseDuration(r.Response.Latency); err != nil {
			log.Errorf("Malformed latency: %v", r.Response.Latency)
			lat = time.Duration(rand.Intn(9)) * time.Second // lat will be between 0 and 9
		}
		// log.Debugf("Introducing latency: %v", lat)
		time.Sleep(lat)
	}
	statusCode = r.Response.Code
	w.WriteHeader(statusCode)
	w.Write(b)

	return statusCode
}

func (r *Route) handleOrdinalRoute(w http.ResponseWriter, req *http.Request) int {
	var b []byte
	var err error

	statusCode := http.StatusInternalServerError
	i := r.index

	if r.Responses == nil {
		w.WriteHeader(statusCode)
		w.Write([]byte("no payloads specified for ordinal endpoint"))
		return statusCode
	}

	if r.Responses[i].Payload != "" {
		b, err = ioutil.ReadFile(r.Responses[i].Payload)
		if err != nil {
			w.WriteHeader(statusCode)
			w.Write([]byte(fmt.Sprintf("failed to read payload: %v", err.Error())))
			return statusCode
		}
	}

	statusCode = r.Responses[i].Code
	if r.Responses[i].Latency != "" {
		var lat time.Duration
		if strings.ToLower(r.Responses[i].Latency) == "random" {
			lat = time.Duration(rand.Intn(9)) * time.Second // lat will be between 0 and 9
		} else if lat, err = time.ParseDuration(r.Responses[i].Latency); err != nil {
			log.Errorf("Malformed latency: %v", r.Responses[i].Latency)
			lat = time.Duration(rand.Intn(9)) * time.Second // lat will be between 0 and 9
		}
		log.Infof("Introducing latency: %v", lat)
		time.Sleep(lat)
	}
	w.WriteHeader(r.Responses[i].Code)
	w.Write(b)

	if i < len(r.Responses)-1 {
		r.index++
	}

	return statusCode
}

func (r *Route) handleVariableRoute(w http.ResponseWriter, req *http.Request) int {
	var b []byte
	var err error

	statusCode := http.StatusInternalServerError
	i := r.index

	if r.Responses == nil {
		w.WriteHeader(statusCode)
		w.Write([]byte("no payloads specified for variable endpoint"))
		return statusCode
	}

	if r.Responses[i].Payload != "" {
		b, err = ioutil.ReadFile(r.Responses[i].Payload)
		if err != nil {
			w.WriteHeader(statusCode)
			w.Write([]byte(fmt.Sprintf("failed to read payload: %v", err.Error())))
			return statusCode
		}
	}
	if r.Responses[i].Latency != "" {
		var lat time.Duration
		if strings.ToLower(r.Responses[i].Latency) == "random" {
			lat = time.Duration(rand.Intn(9)) * time.Second // lat will be between 0 and 9
		} else if lat, err = time.ParseDuration(r.Responses[i].Latency); err != nil {
			log.Errorf("Malformed latency: %v", r.Responses[i].Latency)
			lat = time.Duration(rand.Intn(9)) * time.Second // lat will be between 0 and 9
		}
		log.Infof("Introducing latency: %v", lat)
		time.Sleep(lat)
	}
	statusCode = r.Responses[i].Code
	w.WriteHeader(statusCode)
	w.Write(b)

	return statusCode
}

func (r *Route) handleVerbRoute(w http.ResponseWriter, req *http.Request) int {

	statusCode := http.StatusInternalServerError
	for i := range r.Responses {
		if strings.ToUpper(r.Responses[i].Verb) == req.Method {
			var b []byte
			var err error

			if r.Responses[i].Payload != "" {
				b, err = ioutil.ReadFile(r.Responses[i].Payload)
				if err != nil {
					w.WriteHeader(statusCode)
					w.Write([]byte(fmt.Sprintf("failed to read payload: %v", err.Error())))
					return statusCode
				}
			}
			if r.Responses[i].Latency != "" {
				var lat time.Duration
				if strings.ToLower(r.Responses[i].Latency) == "random" {
					lat = time.Duration(rand.Intn(9)) * time.Second // lat will be between 0 and 9
				} else if lat, err = time.ParseDuration(r.Responses[i].Latency); err != nil {
					log.Errorf("Malformed latency: %v", r.Responses[i].Latency)
					lat = time.Duration(rand.Intn(9)) * time.Second // lat will be between 0 and 9
				}
				log.Infof("Introducing latency: %v", lat)
				time.Sleep(lat)
			}
			statusCode = r.Responses[i].Code
			w.WriteHeader(statusCode)
			w.Write(b)
			return statusCode
		}
	}

	statusCode = http.StatusMethodNotAllowed
	w.WriteHeader(statusCode)
	w.Write([]byte("method not defined in config"))

	return statusCode
}

func (r *Route) handleRateRoute(w http.ResponseWriter, req *http.Request) int {

	idx := -1
	statusCode := http.StatusInternalServerError
	// log.Debugf("Length: %v", len(r.Responses))
	for i := range r.Responses {
		if r.Responses[i].Rate > 0 {
			if rand.Float32() < r.Responses[i].Rate {
				idx = i
			} else {
				if (i + 1) < len(r.Responses) {
					idx = i + 1
				} else {
					idx = 0
				}
			}
			break
		}
	}
	// log.Debugf("idx: %v", idx)
	if idx > -1 {
		var b []byte
		var err error

		if r.Responses[idx].Payload != "" {
			b, err = ioutil.ReadFile(r.Responses[idx].Payload)
			if err != nil {
				w.WriteHeader(statusCode)
				w.Write([]byte(fmt.Sprintf("failed to read payload: %v", err.Error())))
				return statusCode
			}
		}
		if r.Responses[idx].Latency != "" {
			var lat time.Duration
			if strings.ToLower(r.Responses[idx].Latency) == "random" {
				lat = time.Duration(rand.Intn(9)) * time.Second // lat will be between 0 and 9
			} else if lat, err = time.ParseDuration(r.Responses[idx].Latency); err != nil {
				log.Errorf("Malformed latency: %v", r.Responses[idx].Latency)
				lat = time.Duration(rand.Intn(9)) * time.Second // lat will be between 0 and 9
			}
			log.Infof("Introducing latency: %v", lat)
			time.Sleep(lat)
		}
		statusCode = r.Responses[idx].Code
		w.WriteHeader(statusCode)
		w.Write(b)
		return statusCode
	} //if (idx > 1)

	statusCode = http.StatusNotAcceptable
	w.WriteHeader(statusCode)
	w.Write([]byte("rate not defined"))

	return statusCode
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
