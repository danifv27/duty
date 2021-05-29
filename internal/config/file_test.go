package config

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	. "github.com/franela/goblin"
	"github.com/gomicro/ledger"
	"github.com/gomicro/penname"
	. "github.com/onsi/gomega"
)

func TestConfig(t *testing.T) {
	g := Goblin(t)
	RegisterFailHandler(func(m string, _ ...int) { g.Fail(m) })

	g.Describe("Config File", func() {
		g.Before(func() {
			mw := penname.New()
			log = ledger.New(mw, ledger.DebugLevel)
		})

		g.Describe("Parsing from file", func() {
			g.It("should parse a default config file", func() {
				c, err := ParseFromFile()
				Expect(err).To(BeNil())

				Expect(len(c.Routes)).To(Equal(8))
				Expect(c.Status).NotTo(Equal(""))
				Expect(c.Routes[0].Response.Code).To(BeNumerically(">", 0))
				Expect(c.Routes[0].Response.Payload).NotTo(Equal(""))
			})

			g.It("should parse a custom config file set in the environment", func() {
				os.Setenv("DUTY_CONFIG_FILE", "./duty_other.yaml")
				defer os.Unsetenv("DUTY_CONFIG_FILE")

				c, err := ParseFromFile()
				Expect(err).To(BeNil())

				Expect(len(c.Routes)).To(Equal(2))
				Expect(c.Status).NotTo(Equal(""))
				Expect(c.Routes[0].Response.Code).To(BeNumerically(">", 0))
				Expect(c.Routes[0].Response.Payload).NotTo(Equal(""))
			})

			g.It("should return an error when it can't read the file", func() {
				os.Setenv("DUTY_CONFIG_FILE", "./duty_missing.yaml")
				defer os.Unsetenv("DUTY_CONFIG_FILE")

				c, err := ParseFromFile()
				Expect(err).NotTo(BeNil())
				Expect(err.Error()).To(ContainSubstring("Failed to read config file"))
				Expect(c).To(BeNil())
			})
		})

		g.Describe("Status", func() {
			g.It("should serve a default status endpoint", func() {
				f, err := ParseFromFile()
				Expect(err).To(BeNil())

				server := httptest.NewServer(f)
				defer server.Close()

				u := fmt.Sprintf("%v%v", server.URL, "/duty/status")
				res, err := http.Get(u)
				Expect(err).To(BeNil())

				b, err := ioutil.ReadAll(res.Body)
				Expect(err).To(BeNil())
				Expect(string(b)).To(Equal("duty is functioning"))
			})

			g.It("should serve a custom status endpoint", func() {
				f := &File{
					Status: "/another/endpoint",
				}
				server := httptest.NewServer(f)
				defer server.Close()

				u := fmt.Sprintf("%v%v", server.URL, "/another/endpoint")
				res, err := http.Get(u)
				Expect(err).To(BeNil())

				b, err := ioutil.ReadAll(res.Body)
				Expect(err).To(BeNil())
				Expect(string(b)).To(Equal("duty is functioning"))
			})
		})

		g.Describe("Routes", func() {
			g.It("should return a route if it has one", func() {
				f, _ := ParseFromFile()
				u, _ := url.Parse("http://localhost:4567/v1/foo")

				r, found := f.getRoute(u)
				Expect(found).To(BeTrue())
				Expect(r.Endpoint).To(Equal("/v1/foo"))
			})

			g.It("should return false if it doesn't have a route", func() {
				f, _ := ParseFromFile()
				u, _ := url.Parse("http://localhost:4567/v1/notanendpoint")

				r, found := f.getRoute(u)
				Expect(found).To(BeFalse())
				Expect(r).To(BeNil())
			})
		})
	})
}
