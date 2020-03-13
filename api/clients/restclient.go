package clients

import (
	"time"

	"github.com/mercadolibre/golang-restclient/rest"
)

type Client interface {
	Post(string, interface{}) Response
	Put(string, interface{}) Response
	Get(string) Response
	Delete(string) Response
}

type client struct {
	RestClient *rest.RequestBuilder
}

func (c *client) Get(url string) Response {
	r := c.RestClient.Get(url)
	return newResponse(r)
}

func (c *client) Post(url string, body interface{}) Response {
	r := c.RestClient.Post(url, body)
	return newResponse(r)
}

func (c *client) Put(url string, body interface{}) Response {
	r := c.RestClient.Put(url, body)
	return newResponse(r)
}

func (c *client) Delete(url string) Response {
	r := c.RestClient.Delete(url)
	return newResponse(r)
}

type RestClient struct {
	BaseURL     string
	ContentType rest.ContentType
	Timeout     time.Duration
}

type Response interface {
	Bytes() []byte
	Err() error
	StatusCode() int
}

type response struct {
	*rest.Response
}

func newResponse(r *rest.Response) *response {
	return &response{
		Response: r,
	}
}

func (r *response) Bytes() []byte {
	return r.Response.Bytes()
}

func (r *response) Err() error {
	return r.Response.Err
}

func (r *response) StatusCode() int {
	return r.Response.StatusCode
}