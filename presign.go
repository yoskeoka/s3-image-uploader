package main

import (
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

func main() {
	filename := "docker.jpeg"

	h := md5.New()
	content, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer content.Close()

	fi, err := content.Stat()
	if err != nil {
		log.Fatal(err)
	}

	filelen := strconv.Itoa(int(fi.Size()))
	fmt.Println("file length:", filelen)

	if _, err := io.Copy(h, content); err != nil {
		log.Fatal(err)
	}
	content.Seek(0, 0)
	md5s := base64.StdEncoding.EncodeToString(h.Sum(nil))
	fmt.Println("md5 checksum:", md5s)

	// Initialize a session in us-west-2 that the SDK will use to load
	// credentials from the shared credentials file ~/.aws/credentials.
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1")},
	)

	// Create S3 service client
	svc := s3.New(sess)

	resp, _ := svc.PutObjectRequest(&s3.PutObjectInput{
		Bucket:        aws.String("image-uploader-us-east-1-dev"),
		Key:           aws.String("images/" + filename),
		ContentType:   aws.String("image/jpeg"),
		ContentLength: aws.Int64(fi.Size()),
		ACL:           aws.String(s3.BucketCannedACLPublicRead),
	})

	resp.HTTPRequest.Header.Set("Content-MD5", md5s)
	resp.HTTPRequest.Header.Set("Content-Type", "image/jpeg")
	// resp.HTTPRequest.Header.Set("Content-Length", filelen)
	resp.HTTPRequest.ContentLength = fi.Size()

	url, err := resp.Presign(15 * time.Minute)
	if err != nil {
		fmt.Println("error presigning request", err)
		return
	}
	fmt.Println(url)

	req, err := http.NewRequest("PUT", url, content)

	// req.URL.Opaque = strings.Replace(req.URL.Path, ":", "%3A", -1)
	// req.URL.Opaque = strings.Replace(req.URL.Opaque, "|", "%7C", -1)
	// strippedURL := strings.Split(strings.Replace(url, "https://image-uploader-us-east-1-dev.s3.amazonaws.com/", "/", -1), "?")
	// // The actual path
	// req.URL.Opaque = strippedURL[0]
	// // the part with AWS_KEY , Expire, ...
	// req.URL.RawQuery = strippedURL[1]

	req.Header.Set("Content-MD5", md5s)
	req.Header.Set("Content-Type", "image/jpeg")
	req.Header.Set("Content-Length", filelen)
	req.Header.Set("x-amz-acl", s3.BucketCannedACLPublicRead)
	if err != nil {
		fmt.Println("error creating request", url)
		return
	}

	defClient, err := http.DefaultClient.Do(req)
	fmt.Println(defClient, err)
	buf := new(bytes.Buffer)
	buf.ReadFrom(defClient.Body)
	fmt.Println(buf.String())
}
