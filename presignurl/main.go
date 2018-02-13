package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	uuid "github.com/satori/go.uuid"
)

// Response is a response body
type Response struct {
	URL      string `json:"url"`
	Filename string `json:"filename"`
}

// Request is a request body
type Request struct {
	CheckSum string `json:"checksum"`
	Filename string `json:"filename"`
}

// Handler handles GET /presignurl
func Handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	var req Request
	buf := bytes.NewBufferString(request.Body)
	err := json.Unmarshal(buf.Bytes(), &req)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Body:       "must provide md5 checksum of the file",
		}, nil
	}
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(os.Getenv("DEPLOY_REGION"))},
	)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Body:       err.Error(),
		}, nil
	}
	bucketName := os.Getenv("BUCKET_NAME")
	ext := filepath.Ext(req.Filename)
	if ext == "" {
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Body:       "must provide a filename with extention",
		}, nil
	}
	id := uuid.NewV4()
	key := fmt.Sprintf("%v%v", id, ext)

	svc := s3.New(sess)
	r, _ := svc.PutObjectRequest(&s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
	})
	r.HTTPRequest.Header.Set("Content-MD5", req.CheckSum)
	url, err := r.Presign(60 * time.Minute)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Body:       err.Error(),
		}, nil
	}

	// res := Response{URL: url, Filename: key}
	// resJSON, err := json.Marshal(res)
	// if err != nil {
	// 	return events.APIGatewayProxyResponse{
	// 		StatusCode: 500,
	// 	}, nil
	// }
	headers := make(map[string]string)
	headers["Content-Type"] = "text/plain"
	return events.APIGatewayProxyResponse{Headers: headers, Body: url, StatusCode: 200, IsBase64Encoded: false}, nil
}

func main() {
	lambda.Start(Handler)
}
