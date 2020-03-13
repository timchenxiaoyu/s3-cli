package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/urfave/cli"
)

func Del(config *Config, c *cli.Context) error {
	args := c.Args()

	if len(args) != 1 {
		return fmt.Errorf("arguments must be 1: s3 path")
	}

	s3src, err := NewS3Path(args[0])
	if err != nil {
		fmt.Println(err)
		return err
	}

	s3c := GetS3Client(config)

	gobj := &s3.DeleteObjectInput{
		Bucket: aws.String(s3src.Bucket),
		Key:    aws.String(s3src.Path),
	}
	_, err = s3c.DeleteObject(gobj)
	return err
}
