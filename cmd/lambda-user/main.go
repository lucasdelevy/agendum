package main

import (
	"context"
	"encoding/json"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Username  string `json:"username"`
	Email     string `json:"email"`
	Password  string `json:"password"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	UserType  string `json:"userType"`
}

func handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var user User
	if err := json.Unmarshal([]byte(request.Body), &user); err != nil {
		return events.APIGatewayProxyResponse{StatusCode: 400}, err
	}

	sess := session.Must(session.NewSession())
	svc := dynamodb.New(sess)

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return events.APIGatewayProxyResponse{StatusCode: 500}, err
	}

	item := map[string]*dynamodb.AttributeValue{
		"username":  {S: aws.String(user.Username)},
		"email":     {S: aws.String(user.Email)},
		"password":  {S: aws.String(string(hashedPassword))},
		"firstName": {S: aws.String(user.FirstName)},
		"lastName":  {S: aws.String(user.LastName)},
		"userType":  {S: aws.String(user.UserType)},
	}

	_, err = svc.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String(os.Getenv("TABLE_NAME")),
		Item:      item,
	})

	if err != nil {
		return events.APIGatewayProxyResponse{StatusCode: 500}, err
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 201,
		Headers: map[string]string{
			"Content-Type":                 "application/json",
			"Access-Control-Allow-Origin":  "*",
			"Access-Control-Allow-Headers": "Content-Type,X-Amz-Date,Authorization,X-Api-Key,X-Amz-Security-Token",
			"Access-Control-Allow-Methods": "POST,OPTIONS",
		},
		Body: `{"message":"User created successfully"}`,
	}, nil
}

func main() {
	lambda.Start(handler)
}