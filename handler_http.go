package yevna

import (
	"net/http"
)

var DefaultHTTPClient = &HTTPClient{client: http.DefaultClient}

type HTTPClient struct {
	client *http.Client
}

func (h *HTTPClient) SetClient(client *http.Client) *HTTPClient {
	h.client = client
	return h
}

func (h *HTTPClient) Do(fn func(c *Context, in any) (*http.Request, error)) Handler {
	return HandlerFunc(func(c *Context, in any) (any, error) {
		req, err := fn(c, in)
		if err != nil {
			return nil, err
		}

		resp, err := h.client.Do(req)
		if err != nil {
			return nil, err
		}

		out, err := c.Next(resp.Body)

		resp.Body.Close()
		return out, err
	})
}

func HTTP(fn func(c *Context, in any) (*http.Request, error)) Handler {
	return DefaultHTTPClient.Do(fn)
}
