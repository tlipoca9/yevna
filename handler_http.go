package yevna

import (
	"github.com/imroc/req/v3"
)

type HttpHandler struct {
	reqClient *req.Client
	makeReq   func(c *req.Client, in any) *req.Request
}

func HTTP() *HttpHandler {
	return &HttpHandler{
		reqClient: req.C(),
	}
}

func (h *HttpHandler) WithClient(client *req.Client) *HttpHandler {
	h.reqClient = client
	return h
}

func (h *HttpHandler) MakeRequest(fn func(c *req.Client, in any) *req.Request) *HttpHandler {
	h.makeReq = fn
	return h
}

func (h *HttpHandler) Handle(c *Context, in any) (any, error) {
	reqC := h.reqClient.Clone()
	reqC.OnBeforeRequest(func(_ *req.Client, r *req.Request) error {
		return nil
	})

	r := h.makeReq(reqC, in)
	resp := r.Do(c.Context())
	if resp.Err != nil {
		return nil, resp.Err
	}

	out, err := c.Next(resp.Body)

	resp.Body.Close()
	return out, err
}
