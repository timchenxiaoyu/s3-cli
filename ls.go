package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/urfave/cli"
)

func List(config *Config, c *cli.Context) error {
	args := c.Args()
	svc := GetS3Client(config)

	if len(args) == 0 || args[0] == "s3://" {
		var params *s3.ListBucketsInput
		resp, err := svc.ListBuckets(params)
		if err != nil {
			return err
		}

		for _, bucket := range resp.Buckets {
			fmt.Printf("%s  s3://%s\n", bucket.CreationDate.Format(DATE_FMT), *bucket.Name)
		}
		return nil
	}
	return listObject(svc, config, args)
}

func listObject(s3c *s3.S3, config *Config, args []string) error {
	u, err := NewS3Path(args[0])

	if err != nil {
		fmt.Println(err)
		return err
	}
	lo := &s3.ListObjectsInput{
		Bucket:  aws.String(u.Bucket),
		MaxKeys: aws.Int64(1000),
		//Marker:    aws.String(marker),
		Delimiter: aws.String("/"),
		Prefix:    aws.String(u.Path),
	}

	lor, err := s3c.ListObjects(lo)
	if err != nil {
		fmt.Println(err)
		return err
	}

	for _, d := range lor.CommonPrefixes {
		fmt.Printf("%16s %9s   %s://%s/%s\n", "", "DIR", u.Scheme, u.Bucket, *d.Prefix)
	}

	for _, f := range lor.Contents {
		fmt.Printf("%16s %9d   s3://%s/%s\n", f.LastModified.Format(DATE_FMT), *f.Size, u.Bucket, *f.Key)
	}

	return nil
}
