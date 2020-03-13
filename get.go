package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/urfave/cli"
	"io"
	"os"
)

func Get(config *Config, c *cli.Context) error {
	args := c.Args()

	if len(args) != 2 {
		return fmt.Errorf("arguments must be 2: src and dst")
	}

	src, dst := args[0], args[1]

	s3src, err := NewS3Path(src)
	if err != nil {
		fmt.Println(err)
		return err
	}

	s3c := GetS3Client(config)

	dstfile, err := os.Create(dst)
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer dstfile.Close()
	gobj := &s3.GetObjectInput{
		Bucket: aws.String(s3src.Bucket),
		Key:    aws.String(s3src.Path),
	}
	object, err := s3c.GetObject(gobj)
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer object.Body.Close()
	_, err = io.Copy(dstfile, object.Body)

	return err
}
