package main

import (
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/urfave/cli"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strings"
)

func Sync(config *Config, c *cli.Context) error {
	args := c.Args()

	if len(args) != 2 {
		return fmt.Errorf("arguments must be 2: src and dst")
	}

	s, d := args[0], args[1]
	src, serr := NewS3Path(s)

	dst, derr := NewS3Path(d)
	if (serr != nil && derr != nil) || (serr == nil && derr == nil) {

		return errors.New("args must be one is s3 path and another is local path")
	}
	s3c := GetS3Client(config)

	if serr == nil && src.Scheme == "s3" {
		syncS3ToLocal(s3c, src, d)
	} else {
		syncLocalToS3(s3c, s, dst)
	}

	return nil

}

func syncLocalToS3(s3c *s3.S3, src string, s3path *S3Path) {
	s3files := make(map[string]FileInfo)
	if strings.HasSuffix(s3path.Path, "/") {
		s3path.Path = s3path.Path[:len(s3path.Path)-1]
	}
	prefixlen := len(s3path.Path)
	if len(s3path.Path) > 0 {
		prefixlen += 1
	}
	err := getAllS3(s3c, s3path, prefixlen, s3files)
	if err != nil {
		return
	}
	fmt.Println(s3files)

	fmt.Println("+++++++++++++++++++++++")
	localfiles := make(map[string]FileInfo)
	if strings.HasSuffix(src, "/") {
		src = src[:len(src)-1]
	}
	prefixlen = len(src)
	if len(src) > 0 {
		prefixlen += 1
	}
	err = getAllLocal(src, prefixlen, localfiles)
	if err != nil {
		return
	}

	fmt.Println(localfiles)

	createMap, deleteMap := diff(localfiles, s3files)

	createS3File(s3c, s3path, src, createMap)
	deleteS3File(s3c, s3path, deleteMap)
}

func syncS3ToLocal(s3c *s3.S3, s3path *S3Path, dst string) {
	s3files := make(map[string]FileInfo)
	if strings.HasSuffix(s3path.Path, "/") {
		s3path.Path = s3path.Path[:len(s3path.Path)-1]
	}
	prefixlen := len(s3path.Path)
	if len(s3path.Path) > 0 {
		prefixlen += 1
	}
	err := getAllS3(s3c, s3path, prefixlen, s3files)
	if err != nil {
		return
	}
	fmt.Println(s3files)

	fmt.Println("+++++++++++++++++++++++")
	localfiles := make(map[string]FileInfo)
	if strings.HasSuffix(dst, "/") {
		dst = dst[:len(dst)-1]
	}
	prefixlen = len(dst)
	if len(dst) > 0 {
		prefixlen += 1
	}
	err = getAllLocal(dst, prefixlen, localfiles)
	if err != nil {
		return
	}

	fmt.Println(localfiles)

	createMap, deleteMap := diff(s3files, localfiles)

	createLocalFile(s3c, s3path, dst, createMap)
	deleteLocalFile(dst, deleteMap)
}

func getAllLocal(srcdir string, prefixlen int, files map[string]FileInfo) error {

	contents, err := ioutil.ReadDir(srcdir)
	if err != nil {
		fmt.Printf("err read dir %v \n", err)
		return err
	}
	for _, content := range contents {
		//if dir
		if content.IsDir() {
			getAllLocal(path.Join(srcdir, content.Name()), prefixlen, files)
			files[path.Join(srcdir, content.Name())[prefixlen:]] = FileInfo{Type: DIR}
		} else {
			//if file
			etag, err := AmazonEtagHash(path.Join(srcdir, content.Name()))
			if err != nil {
				continue
			}
			files[path.Join(srcdir, content.Name())[prefixlen:]] = FileInfo{Type: FILE, Etag: etag}
		}
	}
	return err
}

func getAllS3(s3c *s3.S3, s3path *S3Path, prefixflen int, s3files map[string]FileInfo) error {

	var mark *string
	for {
		lo := &s3.ListObjectsInput{
			Bucket:    aws.String(s3path.Bucket),
			MaxKeys:   aws.Int64(2),
			Marker:    mark,
			Delimiter: aws.String("/"),
			Prefix:    aws.String(s3path.Path),
		}

		lor, err := s3c.ListObjects(lo)
		if err != nil {
			fmt.Println(err)
			return err
		}

		mark = lor.NextMarker
		for _, d := range lor.CommonPrefixes {
			fmt.Printf("%16s %9s   %s://%s/%s\n", "", "DIR", s3path.Scheme, s3path.Bucket, *d.Prefix)
			s3Pathdir := &S3Path{
				Scheme: s3path.Scheme,
				Bucket: s3path.Bucket,
				Path:   *d.Prefix,
			}
			if len(*d.Prefix) > prefixflen {
				s3files[(*d.Prefix)[prefixflen:len(*d.Prefix)-1]] = FileInfo{Type: DIR}
			}
			getAllS3(s3c, s3Pathdir, prefixflen, s3files)
		}

		for _, f := range lor.Contents {
			fmt.Printf("%16s %9d   s3://%s/%s\n", f.LastModified.Format(DATE_FMT), *f.Size, s3path.Bucket, *f.Key)
			if !strings.HasSuffix(*f.Key, "/") {
				s3files[(*f.Key)[prefixflen:]] = FileInfo{Etag: *f.ETag, Type: FILE}
			}
		}

		if mark == nil {
			return err
		}
	}
	return nil
}

func deleteLocalFile(dst string, deleteMap map[string]FileInfo) {
	for lk, _ := range deleteMap {
		os.RemoveAll(path.Join(dst, lk))
	}
}

func createLocalFile(s3c *s3.S3, s3path *S3Path, dst string, createMap map[string]FileInfo) {

	for ck, cf := range createMap {
		distpath := path.Join(dst, ck)
		os.MkdirAll(path.Dir(distpath), 0777)
		if cf.Type == DIR {
			continue
		}
		dstfile, err := os.Create(distpath)
		if err != nil {
			fmt.Println(err)
			continue
		}
		defer dstfile.Close()
		gobj := &s3.GetObjectInput{
			Bucket: aws.String(s3path.Bucket),
			Key:    aws.String(path.Join(s3path.Path, ck)),
		}
		object, err := s3c.GetObject(gobj)
		if err != nil {
			fmt.Println(err)
			continue
		}
		defer object.Body.Close()
		_, err = io.Copy(dstfile, object.Body)
	}

}

func createS3File(s3c *s3.S3, s3path *S3Path, src string, createMap map[string]FileInfo) {

	uploader := s3manager.NewUploaderWithClient(s3c, func(u *s3manager.Uploader) {
		u.PartSize = 10 * 1024 * 1024
		u.Concurrency = 5
	})
	for ck, cf := range createMap {
		if cf.Type == DIR {
			continue
		}
		fd, err := os.Open(path.Join(src, ck))
		if err != nil {
			fmt.Println(err)
			continue
		}
		defer fd.Close()

		params := &s3manager.UploadInput{
			Bucket: aws.String(s3path.Bucket), // Required
			Key:    aws.String(path.Join(s3path.Path, ck)),
			Body:   fd,
		}

		_, err = uploader.Upload(params)
		if err != nil {
			fmt.Println(err)
			continue
		}
	}

}

func deleteS3File(s3c *s3.S3, s3Path *S3Path, deleteMap map[string]FileInfo) {
	for dk, _ := range deleteMap {
		gobj := &s3.DeleteObjectInput{
			Bucket: aws.String(s3Path.Bucket),
			Key:    aws.String(path.Join(s3Path.Path, dk)),
		}
		_, err := s3c.DeleteObject(gobj)
		if err != nil {
			fmt.Println(err)
			continue
		}
	}

}

func diff(primary, second map[string]FileInfo) (map[string]FileInfo, map[string]FileInfo) {
	//if primary exist ,but second not exist or md5 not equal,add to  create map
	//if primary not exist,but second exist ,add to delete map
	createMap := make(map[string]FileInfo)
	deleteMap := make(map[string]FileInfo)

	for pk, pv := range primary {
		if sv, ok := second[pk]; !ok || sv.Etag != pv.Etag {
			createMap[pk] = pv
		}
		delete(primary, pk)
		delete(second, pk)
	}
	for sk, sv := range second {
		deleteMap[sk] = sv
	}
	return createMap, deleteMap

}
