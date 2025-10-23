package main

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"time"

	"agendum/pkg/auth"
	"agendum/pkg/utils"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

type Team struct {
	Name    string   `json:"name"`
	Admins  []string `json:"admins"`
	Members []string `json:"members"`
}

func handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Check authentication
	token := request.Headers["Authorization"]
	if token == "" {
		return events.APIGatewayProxyResponse{
			StatusCode: 401,
			Headers: map[string]string{
				"Content-Type":                 "application/json",
				"Access-Control-Allow-Origin":  "*",
				"Access-Control-Allow-Headers": "Content-Type,X-Amz-Date,Authorization,X-Api-Key,X-Amz-Security-Token",
				"Access-Control-Allow-Methods": "POST,OPTIONS",
			},
			Body: `{"message":"Authorization header required"}`,
		}, nil
	}

	// Remove "Bearer " prefix if present
	if strings.HasPrefix(token, "Bearer ") {
		token = token[7:]
	}

	_, valid := auth.ValidateToken(token)
	if !valid {
		return events.APIGatewayProxyResponse{
			StatusCode: 401,
			Headers: map[string]string{
				"Content-Type":                 "application/json",
				"Access-Control-Allow-Origin":  "*",
				"Access-Control-Allow-Headers": "Content-Type,X-Amz-Date,Authorization,X-Api-Key,X-Amz-Security-Token",
				"Access-Control-Allow-Methods": "POST,OPTIONS",
			},
			Body: `{"message":"Invalid or expired token"}`,
		}, nil
	}

	var team Team
	if err := json.Unmarshal([]byte(request.Body), &team); err != nil {
		return events.APIGatewayProxyResponse{StatusCode: 400}, err
	}

	sess := session.Must(session.NewSession())
	svc := dynamodb.New(sess)

	teamID := utils.GenerateID()
	createdTimestamp := time.Now().Format(time.RFC3339)

	item := map[string]*dynamodb.AttributeValue{
		"team_id":           {S: aws.String(teamID)},
		"name":              {S: aws.String(team.Name)},
		"created_timestamp": {S: aws.String(createdTimestamp)},
		"admins":            {S: aws.String(strings.Join(team.Admins, ","))},
		"members":           {S: aws.String(strings.Join(team.Members, ","))},
	}

	_, err := svc.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String(os.Getenv("TABLE_NAME")),
		Item:      item,
	})

	if err != nil {
		return events.APIGatewayProxyResponse{StatusCode: 500}, err
	}

	// Update user records to include this team ID
	allUsers := append(team.Admins, team.Members...)
	for _, username := range allUsers {
		_, err := svc.UpdateItem(&dynamodb.UpdateItemInput{
			TableName: aws.String(os.Getenv("USERS_TABLE_NAME")),
			Key: map[string]*dynamodb.AttributeValue{
				"username": {S: aws.String(username)},
			},
			UpdateExpression: aws.String("SET teamIds = list_append(if_not_exists(teamIds, :empty_list), :teamId)"),
			ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
				":teamId": {L: []*dynamodb.AttributeValue{{S: aws.String(teamID)}}},
				":empty_list": {L: []*dynamodb.AttributeValue{}},
			},
		})
		if err != nil {
			// Continue even if user update fails
			continue
		}
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 201,
		Headers: map[string]string{
			"Content-Type":                 "application/json",
			"Access-Control-Allow-Origin":  "*",
			"Access-Control-Allow-Headers": "Content-Type,X-Amz-Date,Authorization,X-Api-Key,X-Amz-Security-Token",
			"Access-Control-Allow-Methods": "POST,OPTIONS",
		},
		Body: `{"message":"Team created successfully","team_id":"` + teamID + `"}`,
	}, nil
}

func main() {
	lambda.Start(handler)
}