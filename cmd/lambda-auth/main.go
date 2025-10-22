package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"golang.org/x/crypto/bcrypt"
)

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Message string `json:"message"`
	Token   string `json:"token"`
}

func generateToken() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return base64.URLEncoding.EncodeToString(bytes)
}

func handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var loginReq LoginRequest
	if err := json.Unmarshal([]byte(request.Body), &loginReq); err != nil {
		return events.APIGatewayProxyResponse{StatusCode: 400}, err
	}

	sess := session.Must(session.NewSession())
	svc := dynamodb.New(sess)

	// Find user by email
	result, err := svc.Scan(&dynamodb.ScanInput{
		TableName: aws.String(os.Getenv("USERS_TABLE_NAME")),
		FilterExpression: aws.String("email = :email"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":email": {S: aws.String(loginReq.Email)},
		},
	})

	if err != nil || len(result.Items) == 0 {
		return events.APIGatewayProxyResponse{
			StatusCode: 401,
			Headers: map[string]string{
				"Content-Type":                 "application/json",
				"Access-Control-Allow-Origin":  "*",
				"Access-Control-Allow-Headers": "Content-Type,X-Amz-Date,Authorization,X-Api-Key,X-Amz-Security-Token",
				"Access-Control-Allow-Methods": "POST,OPTIONS",
			},
			Body: `{"message":"Invalid email or password"}`,
		}, nil
	}

	user := result.Items[0]
	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(*user["password"].S), []byte(loginReq.Password))
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 401,
			Headers: map[string]string{
				"Content-Type":                 "application/json",
				"Access-Control-Allow-Origin":  "*",
				"Access-Control-Allow-Headers": "Content-Type,X-Amz-Date,Authorization,X-Api-Key,X-Amz-Security-Token",
				"Access-Control-Allow-Methods": "POST,OPTIONS",
			},
			Body: `{"message":"Invalid email or password"}`,
		}, nil
	}

	username := *user["username"].S

	// Generate token and store in sessions table
	token := generateToken()
	expiresAt := time.Now().Add(24 * time.Hour).Format(time.RFC3339)

	_, err = svc.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String(os.Getenv("SESSIONS_TABLE_NAME")),
		Item: map[string]*dynamodb.AttributeValue{
			"token":      {S: aws.String(token)},
			"username":   {S: aws.String(username)},
			"expires_at": {S: aws.String(expiresAt)},
		},
	})

	if err != nil {
		return events.APIGatewayProxyResponse{StatusCode: 500}, err
	}

	response := LoginResponse{
		Message: "Login successful",
		Token:   token,
	}

	responseBody, _ := json.Marshal(response)

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers: map[string]string{
			"Content-Type":                 "application/json",
			"Access-Control-Allow-Origin":  "*",
			"Access-Control-Allow-Headers": "Content-Type,X-Amz-Date,Authorization,X-Api-Key,X-Amz-Security-Token",
			"Access-Control-Allow-Methods": "POST,OPTIONS",
		},
		Body: string(responseBody),
	}, nil
}

func main() {
	lambda.Start(handler)
}