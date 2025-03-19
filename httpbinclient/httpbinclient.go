package httpbinclient

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
)

const defaultBaseURL = "http://httpbin.org"

type Client struct {
	httpClient httpClient
	config     Config
}

type httpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type Config struct {
	// BaseURL is the base URL for API requests. Default is <http://httpbin.org>.
	BaseURL *url.URL
}

func NewClient(httpClient httpClient, opts ...Opt) (*Client, error) {
	var config Config
	for _, o := range opts {
		if err := o(&config); err != nil {
			return nil, err
		}
	}

	if config.BaseURL == nil {
		WithBaseURL(defaultBaseURL)(&config)
	}

	return &Client{httpClient: httpClient, config: config}, nil
}

type Opt func(*Config) error

func WithBaseURL(baseURL string) Opt {
	return func(c *Config) error {
		u, err := url.Parse(baseURL)
		if err != nil {
			return err
		}

		c.BaseURL = u

		return nil
	}
}

func (c *Client) Get(ctx context.Context) (*GetResp, error) {
	req, err := http.NewRequest(
		http.MethodGet,
		c.config.BaseURL.JoinPath("/get").String(),
		http.NoBody,
	)
	if err != nil {
		return nil, err
	}

	rawResp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer rawResp.Body.Close()

	resp := GetResp{RawResp: rawResp}
	if err := json.NewDecoder(rawResp.Body).Decode(&resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

type GetResp struct {
	Origin  string `json:"origin"`
	URL     string `json:"url"`
	RawResp *http.Response
}
