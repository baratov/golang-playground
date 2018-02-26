package client

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

type httpClient interface {
	Do(*http.Request) (*http.Response, error)
}

type clientFunc func(*http.Request) (*http.Response, error)

func (c clientFunc) Do(r *http.Request) (*http.Response, error) {
	return c(r)
}

type Middleware func(httpClient) httpClient

func BasicAuthorization(username, password string) Middleware {
	return func(c httpClient) httpClient {
		inner := func(r *http.Request) (*http.Response, error) {
			header := []byte(username + ":" + password)
			encoded := base64.StdEncoding.EncodeToString(header)

			r.Header.Add("Authorization", "Basic "+encoded)
			return c.Do(r)
		}
		return clientFunc(inner)
	}
}

func LogLatency() Middleware {
	return func(c httpClient) httpClient {
		inner := func(r *http.Request) (*http.Response, error) {
			start := time.Now()
			defer func() {
				end := time.Now()
				log.Printf("%v request to %v took %v miliseconds",
					r.Method, r.URL.String(), end.Sub(start).Seconds()*100)
			}()
			return c.Do(r)
		}
		return clientFunc(inner)
	}
}

type Client struct {
	httpClient httpClient
	apiUrl     string
}

const (
	apiVersion = "v1/"
	apiPath    = "api/" + apiVersion + "keys/"
)

func New(apiUrl string, mw ...Middleware) *Client {
	var httpClient httpClient = &http.Client{
		Timeout: time.Second * 10,
	}
	for _, middleware := range mw {
		httpClient = middleware(httpClient)
	}
	return &Client{
		httpClient: httpClient,
		apiUrl:     apiUrl,
	}
}

func (c *Client) Get(key string) (interface{}, error) {
	resp, err := c.makeRequest("GET", key, nil)
	if err != nil {
		return nil, err
	}
	return getValueFromResponse(resp)
}

func (c *Client) Set(key string, value interface{}, ttl time.Duration) error {
	payload, err := c.getPayload(value, ttl)
	if err != nil {
		return err
	}
	resp, err := c.makeRequest("POST", key, payload)
	if err != nil {
		return err
	}
	_, err = getValueFromResponse(resp)
	return err
}

func (c *Client) Update(key string, value interface{}, ttl time.Duration) error {
	payload, err := c.getPayload(value, ttl)
	if err != nil {
		return err
	}
	resp, err := c.makeRequest("PUT", key, payload)
	if err != nil {
		return err
	}
	_, err = getValueFromResponse(resp)
	return err
}

func (c *Client) Delete(key string) error {
	resp, err := c.makeRequest("DELETE", key, nil)
	if err != nil {
		return err
	}
	_, err = getValueFromResponse(resp)
	return err
}

type Payload struct {
	Value interface{}   `json:"value"`
	Ttl   time.Duration `json:"ttl"`
}

func (c *Client) getPayload(value interface{}, ttl time.Duration) (io.Reader, error) {
	p := Payload{Value: value, Ttl: ttl}
	jStr, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}
	return bytes.NewBuffer(jStr), nil
}

func (c *Client) makeRequest(method string, key string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, c.apiUrl+apiPath+key, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	return c.httpClient.Do(req)
}

func getValueFromResponse(r *http.Response) (interface{}, error) {
	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		return nil, err
	}
	var resp map[string]interface{}
	err = json.Unmarshal(body, &resp)
	if err != nil {
		return nil, err
	}
	if msg := resp["message"]; msg != nil {
		return nil, errors.New(msg.(string))
	}
	return resp["data"], nil
}
