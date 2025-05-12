package infra

import (
	"io"
	"net/http"

	"github.com/pkg/errors"
)

type HttpClient struct{}

func NewHttpClient() *HttpClient {
	return &HttpClient{}
}

func (c *HttpClient) HttpGet(url string) (io.ReadCloser, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to make HTTP request")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Wrapf(err, "non-200 response status: %d", resp.StatusCode)
	}

	return resp.Body, nil
}
