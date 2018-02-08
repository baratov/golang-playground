package client

import (
	"encoding/base64"
	"log"
	"net/http"
	"time"
)

type HttpClient interface {
	Do(*http.Request) (*http.Response, error)
}

type ClientFunc func(*http.Request) (*http.Response, error)

func (c ClientFunc) Do(r *http.Request) (*http.Response, error) {
	return c(r)
}

type Middleware func(HttpClient) HttpClient

func BasicAuthorization(username, password string) Middleware {
	return func(c HttpClient) HttpClient {
		inner := func(r *http.Request) (*http.Response, error) {
			header := []byte(username + ":" + password)
			encoded := base64.StdEncoding.EncodeToString(header)

			r.Header.Add("Authorization", "Basic "+encoded)
			return c.Do(r)
		}
		return ClientFunc(inner)
	}
}

func LogResponseTime() Middleware {
	return func(c HttpClient) HttpClient {
		inner := func(r *http.Request) (*http.Response, error) {
			start := time.Now()
			defer func() {
				end := time.Now()
				log.Printf("Request to %s took %v", r.URL.String(), end.Sub(start).Nanoseconds())
			}()
			return c.Do(r)
		}
		return ClientFunc(inner)
	}
}

type Client struct {
	httpClient HttpClient
	apiUrl     string
}

const (
	apiVersion = "v1/"
	apiPath = "api/" + apiVersion + "keys/"
	healthCheckPath = "health/"
)

func New(apiUrl string, mw ...Middleware) *Client {
	var httpClient HttpClient
	for _, middleware := range mw {
		httpClient = middleware(&http.Client{})
	}
	return &Client{
		httpClient: httpClient,
		apiUrl:     apiUrl,
	}
}

func (c *Client) Get(key string) error {

	return nil
}