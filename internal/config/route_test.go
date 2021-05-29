package config

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/franela/goblin"
	. "github.com/onsi/gomega"
)

func TestRoute(t *testing.T) {
	g := Goblin(t)
	RegisterFailHandler(func(m string, _ ...int) { g.Fail(m) })

	g.Describe("Routes", func() {
		g.Describe("CORS", func() {
			g.It("should respond to CORS option requests", func() {
				f, _ := ParseFromFile()

				server := httptest.NewServer(f)
				defer server.Close()

				u := fmt.Sprintf("%v%v", server.URL, "/v1/static")

				c := &http.Client{}

				req, err := http.NewRequest("OPTIONS", u, nil)
				Expect(err).To(BeNil())

				res, err := c.Do(req)
				Expect(err).To(BeNil())
				defer res.Body.Close()

				Expect(res.Header.Get("Access-Control-Allow-Origin")).To(Equal("*"))
				Expect(res.Header.Get("Access-Control-Allow-Methods")).To(Equal("*"))
				Expect(res.Header.Get("Access-Control-Allow-Headers")).To(Equal("*, Authorization"))
				Expect(res.Header.Get("Access-Control-Max-Age")).To(Equal("60"))
				Expect(res.Header.Get("Cache-Control")).To(Equal("no-store, no-cache, must-revalidate, post-check=0, pre-check=0"))
				Expect(res.Header.Get("Vary")).To(Equal("Accept-Encoding"))

				Expect(res.StatusCode).To(Equal(http.StatusNoContent))
			})
		})

		g.Describe("Static", func() {
			g.It("should return static content for a static route", func() {
				f, _ := ParseFromFile()

				server := httptest.NewServer(f)
				defer server.Close()

				u := fmt.Sprintf("%v%v", server.URL, "/v1/static")

				res, err := http.Get(u)
				Expect(err).To(BeNil())
				defer res.Body.Close()

				firstb, err := ioutil.ReadAll(res.Body)
				Expect(err).To(BeNil())
				Expect(string(firstb)).To(ContainSubstring("here lies a foo"))
				Expect(string(firstb)).To(ContainSubstring("git to the bar"))
				Expect(string(firstb)).To(ContainSubstring("let's get the baz back together"))

				res, err = http.Get(u)
				Expect(err).To(BeNil())
				defer res.Body.Close()

				secondb, err := ioutil.ReadAll(res.Body)
				Expect(err).To(BeNil())
				Expect(string(secondb)).To(ContainSubstring("here lies a foo"))
				Expect(string(secondb)).To(ContainSubstring("git to the bar"))
				Expect(string(secondb)).To(ContainSubstring("let's get the baz back together"))

				Expect(firstb).To(Equal(secondb))
			})

			g.It("should treat a route as a static if no type is specified", func() {
				f, _ := ParseFromFile()

				server := httptest.NewServer(f)
				defer server.Close()

				u := fmt.Sprintf("%v%v", server.URL, "/v1/foo")

				res, err := http.Get(u)
				Expect(err).To(BeNil())
				defer res.Body.Close()

				firstb, err := ioutil.ReadAll(res.Body)
				Expect(err).To(BeNil())
				Expect(string(firstb)).To(ContainSubstring("here lies a foo"))
				Expect(string(firstb)).To(ContainSubstring("git to the bar"))
				Expect(string(firstb)).To(ContainSubstring("let's get the baz back together"))

				res, err = http.Get(u)
				Expect(err).To(BeNil())
				defer res.Body.Close()

				secondb, err := ioutil.ReadAll(res.Body)
				Expect(err).To(BeNil())
				Expect(string(secondb)).To(ContainSubstring("here lies a foo"))
				Expect(string(secondb)).To(ContainSubstring("git to the bar"))
				Expect(string(secondb)).To(ContainSubstring("let's get the baz back together"))

				Expect(firstb).To(Equal(secondb))
			})

			g.It("should return a response code for an endpoint with no response defined", func() {
				f, _ := ParseFromFile()

				server := httptest.NewServer(f)
				defer server.Close()

				u := fmt.Sprintf("%v%v", server.URL, "/v1/baz")

				res, err := http.Get(u)
				Expect(err).To(BeNil())
				defer res.Body.Close()

				Expect(res.StatusCode).To(Equal(401))

				u = fmt.Sprintf("%v%v", server.URL, "/v1/biz")

				res, err = http.Get(u)
				Expect(err).To(BeNil())
				defer res.Body.Close()

				Expect(res.StatusCode).To(Equal(200))
			})
		})

		g.Describe("Oridinal", func() {
			g.It("should return ordinal content for an ordinal route", func() {
				f, _ := ParseFromFile()

				server := httptest.NewServer(f)
				defer server.Close()

				u := fmt.Sprintf("%v%v", server.URL, "/v1/ordinal")

				res, err := http.Get(u)
				Expect(err).To(BeNil())
				defer res.Body.Close()

				firstb, err := ioutil.ReadAll(res.Body)
				Expect(err).To(BeNil())
				Expect(string(firstb)).To(ContainSubstring("here lies a foo"))
				Expect(string(firstb)).To(ContainSubstring("git to the bar"))
				Expect(string(firstb)).To(ContainSubstring("let's get the baz back together"))

				res, err = http.Get(u)
				Expect(err).To(BeNil())
				defer res.Body.Close()

				secondb, err := ioutil.ReadAll(res.Body)
				Expect(err).To(BeNil())
				Expect(string(secondb)).To(ContainSubstring("unauthorized"))

				Expect(string(firstb)).NotTo(Equal(string(secondb)))

				res, err = http.Get(u)
				Expect(err).To(BeNil())
				defer res.Body.Close()

				thirdb, err := ioutil.ReadAll(res.Body)
				Expect(err).To(BeNil())
				Expect(string(secondb)).To(ContainSubstring("unauthorized"))

				Expect(string(firstb)).NotTo(Equal(string(thirdb)))
				Expect(string(secondb)).To(Equal(string(thirdb)))
			})

			g.It("should reset an ordinal route", func() {
				f, _ := ParseFromFile()

				server := httptest.NewServer(f)
				defer server.Close()

				u := fmt.Sprintf("%v%v", server.URL, "/v1/ordinal")

				res, err := http.Get(u)
				Expect(err).To(BeNil())
				defer res.Body.Close()

				firstb, err := ioutil.ReadAll(res.Body)
				Expect(err).To(BeNil())
				Expect(string(firstb)).To(ContainSubstring("here lies a foo"))
				Expect(string(firstb)).To(ContainSubstring("git to the bar"))
				Expect(string(firstb)).To(ContainSubstring("let's get the baz back together"))

				res, err = http.Get(u)
				Expect(err).To(BeNil())
				defer res.Body.Close()

				secondb, err := ioutil.ReadAll(res.Body)
				Expect(err).To(BeNil())
				Expect(string(secondb)).To(ContainSubstring("unauthorized"))

				Expect(string(firstb)).NotTo(Equal(string(secondb)))

				r := fmt.Sprintf("%v%v", server.URL, "/duty/reset")
				http.Get(r)

				res, err = http.Get(u)
				Expect(err).To(BeNil())
				defer res.Body.Close()

				thirdb, err := ioutil.ReadAll(res.Body)
				Expect(err).To(BeNil())
				Expect(string(thirdb)).To(ContainSubstring("here lies a foo"))
				Expect(string(thirdb)).To(ContainSubstring("git to the bar"))
				Expect(string(thirdb)).To(ContainSubstring("let's get the baz back together"))

				Expect(string(firstb)).To(Equal(string(thirdb)))
				Expect(string(secondb)).NotTo(Equal(string(thirdb)))

				res, err = http.Get(u)
				Expect(err).To(BeNil())
				defer res.Body.Close()

				fourthb, err := ioutil.ReadAll(res.Body)
				Expect(err).To(BeNil())
				Expect(string(fourthb)).To(ContainSubstring("unauthorized"))

				Expect(string(firstb)).NotTo(Equal(string(fourthb)))
				Expect(string(secondb)).To(Equal(string(fourthb)))
				Expect(string(thirdb)).NotTo(Equal(string(fourthb)))
			})
		})

		g.Describe("Variable", func() {
			g.It("should return variable content for an variable route", func() {
				f, _ := ParseFromFile()

				server := httptest.NewServer(f)
				defer server.Close()

				u := fmt.Sprintf("%v%v", server.URL, "/v1/variable")

				res, err := http.Get(u)
				Expect(err).To(BeNil())
				defer res.Body.Close()

				firstb, err := ioutil.ReadAll(res.Body)
				Expect(err).To(BeNil())
				Expect(string(firstb)).To(ContainSubstring("here lies a foo"))
				Expect(string(firstb)).To(ContainSubstring("git to the bar"))
				Expect(string(firstb)).To(ContainSubstring("let's get the baz back together"))

			})

			g.It("should not increment index each time", func() {
				f, _ := ParseFromFile()

				server := httptest.NewServer(f)
				defer server.Close()

				u := fmt.Sprintf("%v%v", server.URL, "/v1/variable")

				res, err := http.Get(u)
				Expect(err).To(BeNil())
				defer res.Body.Close()

				res, err = http.Get(u)
				Expect(err).To(BeNil())
				defer res.Body.Close()

				firstb, err := ioutil.ReadAll(res.Body)
				Expect(err).To(BeNil())
				Expect(string(firstb)).To(ContainSubstring("here lies a foo"))
				Expect(string(firstb)).To(ContainSubstring("git to the bar"))
				Expect(string(firstb)).To(ContainSubstring("let's get the baz back together"))
			})

			g.It("should set a variable route to id 404 and not reset after", func() {
				f, _ := ParseFromFile()

				server := httptest.NewServer(f)
				defer server.Close()

				r := fmt.Sprintf("%v%v", server.URL, "/duty/set?name=var&id=404")
				res, err := http.Get(r)
				Expect(err).To(BeNil())
				defer res.Body.Close()

				u := fmt.Sprintf("%v%v", server.URL, "/v1/variable")
				res, err = http.Get(u)
				Expect(err).To(BeNil())
				defer res.Body.Close()
				Expect(res.StatusCode).To(Equal(404))

				u = fmt.Sprintf("%v%v", server.URL, "/v1/variable")
				res, err = http.Get(u)
				Expect(err).To(BeNil())
				defer res.Body.Close()
				Expect(res.StatusCode).To(Equal(404))

			})

			g.It("should set a variable route to id 401", func() {
				f, _ := ParseFromFile()

				server := httptest.NewServer(f)
				defer server.Close()

				r := fmt.Sprintf("%v%v", server.URL, "/duty/set?name=var&id=401")
				res, err := http.Get(r)
				Expect(err).To(BeNil())
				defer res.Body.Close()

				u := fmt.Sprintf("%v%v", server.URL, "/v1/variable")
				res, err = http.Get(u)
				Expect(err).To(BeNil())
				defer res.Body.Close()
				Expect(res.StatusCode).To(Equal(401))
			})
		})

		g.Describe("Verb", func() {
			var f *File
			var server *httptest.Server
			var u string

			g.BeforeEach(func() {
				f, _ = ParseFromFile()
				server = httptest.NewServer(f)
				u = fmt.Sprintf("%v%v", server.URL, "/v1/verb")
			})

			g.AfterEach(func() {
				server.Close()
			})

			g.It("should return content based on verb", func() {
				res, err := http.Get(u)
				Expect(err).To(BeNil())
				defer res.Body.Close()

				firstb, err := ioutil.ReadAll(res.Body)
				Expect(err).To(BeNil())
				Expect(string(firstb)).To(ContainSubstring("here lies a foo"))
				Expect(string(firstb)).To(ContainSubstring("git to the bar"))
				Expect(string(firstb)).To(ContainSubstring("let's get the baz back together"))

				res, err = http.Post(u, "application/json", nil)
				Expect(err).To(BeNil())
				defer res.Body.Close()

				secondb, err := ioutil.ReadAll(res.Body)
				Expect(err).To(BeNil())
				Expect(firstb).NotTo(Equal(secondb))
				Expect(string(secondb)).To(ContainSubstring("new foo"))
				Expect(string(secondb)).To(ContainSubstring("here lies a foo"))
				Expect(string(secondb)).To(ContainSubstring("new bar"))
				Expect(string(secondb)).To(ContainSubstring("git to the bar"))
				Expect(string(secondb)).To(ContainSubstring("new baz"))
				Expect(string(secondb)).To(ContainSubstring("let's get the baz back together"))
			})

			g.It("should return a 405 on a verb not matched", func() {
				res, err := http.Head(u)
				Expect(err).To(BeNil())
				defer res.Body.Close()

				Expect(res.StatusCode).To(Equal(405))
			})
		})
	})
}
