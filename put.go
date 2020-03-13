package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/urfave/cli"
	"os"
	"path"
)

func Put(config *Config, c *cli.Context) error {
	args := c.Args()

	if len(args) < 2 {
		return fmt.Errorf("Not enought arguments")
	}

	src, dst := args[:len(args)-1], args[len(args)-1]

	s3dst, err := NewS3Path(dst)
	if err != nil {
		fmt.Println(err)
		return err
	}

	svc := GetS3Client(config)

	uploader := s3manager.NewUploaderWithClient(svc, func(u *s3manager.Uploader) {
		u.PartSize = 10 * 1024 * 1024
		u.Concurrency = 5
	})
	for _, srcfile := range src {
		fd, err := os.Open(srcfile)
		if err != nil {
			return err
		}
		defer fd.Close()

		if len(s3dst.Path) == 0 {
			s3dst.Path = path.Base(srcfile)
		}

		params := &s3manager.UploadInput{
			Bucket: aws.String(s3dst.Bucket), // Required
			Key:    aws.String(s3dst.Path),
			Body:   fd,
		}

		_, err = uploader.Upload(params)
		if err != nil {
			return err
		}
	}

	return nil
}
