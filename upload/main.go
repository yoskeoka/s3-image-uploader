package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	uuid "github.com/satori/go.uuid"
)

// Response is a body of /upload response
type Response struct {
	Message string `json:"message"`
}

// NewResJSON create a response json
func NewResJSON(msg string) string {
	res := Response{msg}
	resJSON, err := json.Marshal(&res)
	if err != nil {
		panic(err)
	}
	return string(resJSON)
}

var mime = map[string]string{
	"image/png":  ".png",
	"image/jpeg": ".jpg",
}

// Request is a body of /upload request
type Request struct {
	MIMEType string `json:"mime_type"`
	Content  string `json:"content"`
}

// Handler handles POST /upload
func Handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	region := os.Getenv("DEPLOY_REGION")
	bucketName := os.Getenv("BUCKET_NAME")
	sizeLimitMBStr := os.Getenv("UPLOAD_SIZE_LIMIT_MB")
	sizeLimitMB, err := strconv.Atoi(sizeLimitMBStr)
	if err != nil {
		panic(err)
	}

	var body Request
	buf := bytes.NewBufferString(request.Body)
	err = json.Unmarshal(buf.Bytes(), &body)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Body:       NewResJSON("request body mismatch"),
		}, nil
	}

	ext, ok := mime[body.MIMEType]
	if !ok {
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Body:       NewResJSON(fmt.Sprintf("mime_type %v is not allowed", body.MIMEType)),
		}, nil
	}

	imgBuf, err := base64.StdEncoding.DecodeString(body.Content)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Body:       NewResJSON(fmt.Sprintf("content must be encoded base64: %v", err.Error())),
		}, nil
	}

	uid := uuid.NewV4()
	key := "images/" + uid.String() + ext
	content := bytes.NewReader(imgBuf)

	fmt.Println("file length:", len(imgBuf))
	sizeLimit := sizeLimitMB * (1 << (10 * 2))
	if len(imgBuf) > sizeLimit {
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Body:       NewResJSON(fmt.Sprintf("content length exceeds the limit. (%vMB)", sizeLimitMB)),
		}, nil
	}

	// Initialize a session in us-west-2 that the SDK will use to load
	// credentials from the shared credentials file ~/.aws/credentials.
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(region)},
	)

	// Create S3 service client
	svc := s3.New(sess)

	_, err = svc.PutObject(&s3.PutObjectInput{
		Body:          content,
		Bucket:        aws.String(bucketName),
		Key:           aws.String(key),
		ContentType:   aws.String(body.MIMEType),
		ContentLength: aws.Int64(int64(content.Len())),
		ACL:           aws.String(s3.BucketCannedACLPublicRead),
	})

	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Body:       NewResJSON(err.Error()),
		}, nil
	}
	url := fmt.Sprintf("http://%v.s3-website-%v.amazonaws.com/%v", bucketName, region, key)
	return events.APIGatewayProxyResponse{Body: NewResJSON(url), StatusCode: 200}, nil
}

func main() {
	lambda.Start(Handler)
}
