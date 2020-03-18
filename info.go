package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/urfave/cli"
	"path"
)

func Info(config *Config, c *cli.Context) error {
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

	if len(s3path.Path) == 0 {
		return infoBucket(s3c, s3path)
	}
	return infoObject(s3c, s3path)
}

func infoObject(s3c *s3.S3, s3path *S3Path) error {
	//head object
	hoi := &s3.HeadObjectInput{
		Bucket: aws.String(s3path.Bucket),
		Key:    aws.String(s3path.Path),
	}
	hoir, err := s3c.HeadObject(hoi)
	if err != nil {
		fmt.Println(err)
		return err
	}
	// bucket policy
	gbpi := &s3.GetBucketPolicyInput{
		Bucket: aws.String(s3path.Bucket),
	}
	gbpr, err := s3c.GetBucketPolicy(gbpi)
	if err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Printf("%s://%s (object): \n", s3path.Scheme, path.Join(s3path.Bucket, s3path.Path))
	fmt.Printf("    File size: %d\n", *hoir.ContentLength)
	fmt.Printf("    Last mod: %s\n", *hoir.LastModified)
	fmt.Printf("    MIME type: %s\n", *hoir.ContentType)
	fmt.Printf("    MD5 sum: %s\n", *hoir.ETag)
	fmt.Printf("    Policy: %s\n", *gbpr.Policy)

	return nil
}

func infoBucket(s3c *s3.S3, s3path *S3Path) error {

	//bucket policy
	gbpi := &s3.GetBucketPolicyInput{
		Bucket: aws.String(s3path.Bucket),
	}
	gbpr, err := s3c.GetBucketPolicy(gbpi)
	if err != nil {
		fmt.Println(err)
		return err
	}

	//bucket location
	gbli := &s3.GetBucketLocationInput{
		Bucket: aws.String(s3path.Bucket),
	}
	gblr, err := s3c.GetBucketLocation(gbli)
	if err != nil {
		fmt.Println(err)
		return err
	}
	//bucket acl
	gbai := &s3.GetBucketAclInput{
		Bucket: aws.String(s3path.Bucket),
	}
	gbar, err := s3c.GetBucketAcl(gbai)
	if err != nil {
		fmt.Println(err)
		return err
	}

	fmt.Printf("%s://%s (bucket): \n", s3path.Scheme, s3path.Bucket)
	fmt.Printf("    Location: %s\n", *gblr.LocationConstraint)
	fmt.Printf("    Policy: %s\n", *gbpr.Policy)
	if len(gbar.Grants) > 0 {
		fmt.Printf("    ACL: %9s:%s\n", *gbar.Grants[0].Grantee.ID, *gbar.Grants[0].Permission)
	}

	return nil
}
