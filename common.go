package main

import (
	"crypto/md5"
	"fmt"
	"io"
	"net/url"
	"os"
)

const DATE_FMT = "2006-01-02 15:04"

type FileType string

const (
	BYTE = 1 << (10 * iota)
	KB
	MB
	GB
	TB
)

const (
	DIR  FileType = "DIR"
	FILE FileType = "FILE"
)

type S3Path struct {
	Scheme string
	Bucket string
	Path   string
}

type FileInfo struct {
	Path string
	Etag string
	Type FileType
}

func NewS3Path(path string) (*S3Path, error) {
	u, err := url.Parse(path)
	if err != nil {
		return nil, err
	}
	if u.Scheme != "s3" {
		return nil, fmt.Errorf("Invalid URI scheme must be  s3://")
	}

	uri := S3Path{
		Scheme: u.Scheme,
		Bucket: u.Host,
		Path:   u.Path,
	}

	if uri.Scheme == "s3" && uri.Path != "" {
		uri.Path = uri.Path[1:]
	}

	return &uri, nil
}

func AmazonEtagHash(path string) (string, error) {
	const BLOCK_SIZE = 1024 * 1024 * 5    // 5MB
	const START_BLOCKS = 1024 * 1024 * 16 // 16MB

	fd, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer fd.Close()

	info, err := fd.Stat()
	if err != nil {
		return "", err
	}

	hasher := md5.New()
	count := 0

	if info.Size() > START_BLOCKS {
		for err != io.EOF {
			count += 1
			parthasher := md5.New()
			var size int64
			size, err = io.CopyN(parthasher, fd, BLOCK_SIZE)
			if err != nil && err != io.EOF {
				return "", err
			}
			if size != 0 {
				hasher.Write(parthasher.Sum(nil))
			}
		}
	} else {
		if _, err := io.Copy(hasher, fd); err != nil {
			return "", err
		}
	}

	hash := fmt.Sprintf("%x", hasher.Sum(nil))

	if count != 0 {
		hash += fmt.Sprintf("-%d", count)
	}
	return hash, nil
}

func HumanReadByteSize(size float64) string {
	if size > TB {
		return fmt.Sprintf("%fT", size/TB)
	} else if size > GB {
		return fmt.Sprintf("%fG", size/GB)
	} else if size > MB {
		return fmt.Sprintf("%fM", size/MB)
	} else if size > KB {
		return fmt.Sprintf("%fKB", size/KB)
	}

	return fmt.Sprintf("%fB", size)
}
