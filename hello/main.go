package main

import (
	"encoding/json"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

// Response is a response body
type Response struct {
	Message string `json:"message"`
}

// Handler handles GET /hello
func Handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	res := Response{
		Message: "Go Serverless v1.0! Your function executed successfully!"}
	resJSON, err := json.Marshal(res)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
		}, nil
	}
	return events.APIGatewayProxyResponse{Body: string(resJSON), StatusCode: 200}, nil
}

func main() {
	lambda.Start(Handler)
}
