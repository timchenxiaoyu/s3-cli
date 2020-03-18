package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/urfave/cli"
)

func Du(config *Config, c *cli.Context) error {
	args := c.Args()
	s3c := GetS3Client(config)

	if len(args) != 1 {
		return fmt.Errorf("arguments must be 1 bucket path")
	}
	s3path, err := NewS3Path(args[0])

	if err != nil {
		fmt.Println(err)
		return err
	}
	var ObjectCount int64
	var ObjectSize int64
	var mark *string
	for {
		lo := &s3.ListObjectsInput{
			Bucket:  aws.String(s3path.Bucket),
			MaxKeys: aws.Int64(1000),
			Marker:  mark,
			//Delimiter: aws.String("/"),
			Prefix: aws.String(s3path.Path),
		}

		lor, err := s3c.ListObjects(lo)
		if err != nil {
			fmt.Println(err)
			return err
		}

		mark = lor.NextMarker

		for _, f := range lor.Contents {
			ObjectCount += 1
			ObjectSize += *f.Size
		}

		if mark == nil {
			break
		}
	}
	fmt.Printf("%s %4d objects s3://%s/%s\n", HumanReadByteSize(float64(ObjectSize)), ObjectCount, s3path.Bucket, s3path.Path)

	return nil

}
