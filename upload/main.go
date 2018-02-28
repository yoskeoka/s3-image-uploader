package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	uuid "github.com/satori/go.uuid"
)

// ErrorResponse is a body of /upload error response
type ErrorResponse struct {
	Message string `json:"message"`
}

// NewErrorResJSON create a response json
func NewErrorResJSON(msg string) string {
	res := ErrorResponse{msg}
	resJSON, err := json.Marshal(res)
	if err != nil {
		panic(err)
	}
	return string(resJSON)
}

// SuccessResponse is a success response
type SuccessResponse struct {
	URL string `json:"url"`
}

var mime = map[string]string{
	"image/png":  ".png",
	"image/jpeg": ".jpg",
}

var headers = map[string]string{
	"Content-Type":                "application/json",
	"Access-Control-Allow-Origin": "*",
}

// Request is a body of /upload request
type Request struct {
	MIMEType string `json:"mime_type"`
	Content  string `json:"content"`
}

// Handler handles POST /upload
func Handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	region := os.Getenv("AWS_REGION")
	bucketName := os.Getenv("BUCKET_NAME")
	bucketURL := os.Getenv("BUCKET_URL")
	subDir := os.Getenv("SUB_DIR")
	appendDateStr := os.Getenv("APPEND_DATE")
	appendDate := false
	if appendDateStr == "true" {
		appendDate = true
	}

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
			Headers:    headers,
			Body:       NewErrorResJSON("request body mismatch"),
		}, nil
	}

	ext, ok := mime[body.MIMEType]
	if !ok {
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Headers:    headers,
			Body:       NewErrorResJSON(fmt.Sprintf("mime_type %v is not allowed", body.MIMEType)),
		}, nil
	}

	imgBuf, err := base64.StdEncoding.DecodeString(body.Content)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Headers:    headers,
			Body:       NewErrorResJSON(fmt.Sprintf("content must be encoded base64: %v", err.Error())),
		}, nil
	}

	content := bytes.NewReader(imgBuf)

	fmt.Println("file length:", len(imgBuf))
	sizeLimit := sizeLimitMB * (1 << (10 * 2))
	if len(imgBuf) > sizeLimit {
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Headers:    headers,
			Body:       NewErrorResJSON(fmt.Sprintf("content length exceeds the limit. (%vMB)", sizeLimitMB)),
		}, nil
	}

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(region)},
	)

	// Create S3 service client
	svc := s3.New(sess)

	// build S3 ObjectKey
	uid := uuid.NewV4()

	var key string
	if appendDate {
		// add date key for auto removable
		date := time.Now().UTC().Format("20060102") //yyyyMMdd
		key = fmt.Sprintf("%s/%s/%s%s", subDir, date, uid.String(), ext)
	} else {
		key = fmt.Sprintf("%s/%s%s", subDir, uid.String(), ext)
	}

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
			Headers:    headers,
			Body:       NewErrorResJSON(err.Error()),
		}, nil
	}
	url := fmt.Sprintf("%v/%v", bucketURL, key)
	res := SuccessResponse{url}
	resJSON, err := json.Marshal(res)
	if err != nil {
		panic(err)
	}

	// resize
	risizeDomain := os.Getenv("RISIZE_DOMAIN")
	if risizeDomain != "" {
		if err := resize(bucketName, key); err != nil {
			panic(err)
		}
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers:    headers,
		Body:       string(resJSON),
	}, nil
}

// resize
func resize(bucketName string, key string) error {

	risizeURL := os.Getenv("RISIZE_URL")
	jsonStr := fmt.Sprintf(`{"branch_name":"%s", "image_file_path": "%s", "size": {"width": 387, "height": 387}, "quality": 80}`, bucketName, key)

	req, err := http.NewRequest(
		"POST",
		risizeURL,
		bytes.NewBuffer([]byte(jsonStr)),
	)
	if err != nil {
		return err
	}

	// Content-Type 設定
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return err
}

func main() {
	lambda.Start(Handler)
}
