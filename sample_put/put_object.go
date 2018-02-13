package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

func main() {
	filename := "docker.jpeg"

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

	// Initialize a session in us-west-2 that the SDK will use to load
	// credentials from the shared credentials file ~/.aws/credentials.
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1")},
	)

	// Create S3 service client
	svc := s3.New(sess)

	_, err = svc.PutObject(&s3.PutObjectInput{
		Body:          content,
		Bucket:        aws.String("image-uploader-us-east-1-dev"),
		Key:           aws.String("direct_put_images/" + filename),
		ContentType:   aws.String("image/jpeg"),
		ContentLength: aws.Int64(fi.Size()),
		ACL:           aws.String(s3.BucketCannedACLPublicRead),
	})

	if err != nil {
		log.Fatal(err.Error())
	}

}
