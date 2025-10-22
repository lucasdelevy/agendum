package main

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"agendum/pkg/utils"
)

type Team struct {
	Name    string   `json:"name"`
	Admins  []string `json:"admins"`
	Members []string `json:"members"`
}

func handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
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

	return events.APIGatewayProxyResponse{
		StatusCode: 201,
		Headers:    map[string]string{"Content-Type": "application/json"},
		Body:       `{"message":"Team created successfully","team_id":"` + teamID + `"}`,
	}, nil
}

func main() {
	lambda.Start(handler)
}