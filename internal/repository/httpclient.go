package repository

import "io"

type HttpClient interface {
	Get(url string) (io.ReadCloser, error)
}
