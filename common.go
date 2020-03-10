package main

import (
	"fmt"
	"net/url"
)

const DATE_FMT = "2006-01-02 15:04"

type S3Path struct {
	Scheme string
	Bucket string
	Path   string
}

func FileURINew(path string) (*FileURI, error) {
	u, err := url.Parse(path)
	if err != nil {
		return nil, err
	}
	if u.Scheme != "" && u.Scheme != "s3" && u.Scheme != "file" {
		return nil, fmt.Errorf("Invalid URI scheme must be one of file/s3/NONE")
	}

	uri := FileURI{
		Scheme: u.Scheme,
		Bucket: u.Host,
		Path:   u.Path,
	}

	if uri.Scheme == "" {
		uri.Scheme = "file"
	}
	if uri.Scheme == "s3" && uri.Path != "" {
		uri.Path = uri.Path[1:]
	}
	if uri.Path == "" && uri.Scheme == "s3" {
		uri.Path = "/"
	}

	return &uri, nil
}
