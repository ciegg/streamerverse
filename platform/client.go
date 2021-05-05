package platform

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

type Client struct {
	wrapper http.Client
}

func NewClient() *Client {
	return &Client{
		http.Client{Timeout: 30 * time.Second},
	}
}

func (c *Client) Post(url string, headers http.Header, body io.Reader) (*http.Response, error) {
	return c.do(http.MethodPost, url, headers, body)
}

func (c *Client) Get(url string, headers http.Header) (*http.Response, error) {
	return c.do(http.MethodGet, url, headers, nil)
}

func (c *Client) do(method, url string, headers http.Header, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	req.Header = headers

	resp, err := c.wrapper.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("Non-OK HTTP status: %s", resp.Status)
	}

	return resp, nil
}
