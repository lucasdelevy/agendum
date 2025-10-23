package main

import (
	"context"
	"encoding/json"
	"os"
	"strings"

	"agendum/pkg/auth"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

type Team struct {
	TeamID  string   `json:"team_id"`
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
				"Access-Control-Allow-Methods": "GET,OPTIONS",
			},
			Body: `{"message":"Authorization header required"}`,
		}, nil
	}

	// Remove "Bearer " prefix if present
	if strings.HasPrefix(token, "Bearer ") {
		token = token[7:]
	}

	username, valid := auth.ValidateToken(token)
	if !valid {
		return events.APIGatewayProxyResponse{
			StatusCode: 401,
			Headers: map[string]string{
				"Content-Type":                 "application/json",
				"Access-Control-Allow-Origin":  "*",
				"Access-Control-Allow-Headers": "Content-Type,X-Amz-Date,Authorization,X-Api-Key,X-Amz-Security-Token",
				"Access-Control-Allow-Methods": "GET,OPTIONS",
			},
			Body: `{"message":"Invalid or expired token"}`,
		}, nil
	}

	sess := session.Must(session.NewSession())
	svc := dynamodb.New(sess)

	// Get user's team IDs
	userResult, err := svc.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String(os.Getenv("USERS_TABLE_NAME")),
		Key: map[string]*dynamodb.AttributeValue{
			"username": {S: aws.String(username)},
		},
	})
	if err != nil {
		return events.APIGatewayProxyResponse{StatusCode: 500}, err
	}

	if userResult.Item == nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 404,
			Headers: map[string]string{
				"Content-Type":                 "application/json",
				"Access-Control-Allow-Origin":  "*",
				"Access-Control-Allow-Headers": "Content-Type,X-Amz-Date,Authorization,X-Api-Key,X-Amz-Security-Token",
				"Access-Control-Allow-Methods": "GET,OPTIONS",
			},
			Body: `{"message":"User not found"}`,
		}, nil
	}

	var teamIDs []string
	if teamIDsAttr, exists := userResult.Item["teamIds"]; exists && teamIDsAttr.L != nil {
		for _, teamID := range teamIDsAttr.L {
			if teamID.S != nil {
				teamIDs = append(teamIDs, *teamID.S)
			}
		}
	}

	var teams []Team
	for _, teamID := range teamIDs {
		teamResult, err := svc.GetItem(&dynamodb.GetItemInput{
			TableName: aws.String(os.Getenv("TEAMS_TABLE_NAME")),
			Key: map[string]*dynamodb.AttributeValue{
				"team_id": {S: aws.String(teamID)},
			},
		})
		if err != nil {
			continue
		}

		if teamResult.Item != nil {
			team := Team{
				TeamID: *teamResult.Item["team_id"].S,
				Name:   *teamResult.Item["name"].S,
			}
			if teamResult.Item["admins"].S != nil {
				team.Admins = strings.Split(*teamResult.Item["admins"].S, ",")
			}
			if teamResult.Item["members"].S != nil {
				team.Members = strings.Split(*teamResult.Item["members"].S, ",")
			}
			teams = append(teams, team)
		}
	}

	response, _ := json.Marshal(teams)
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers: map[string]string{
			"Content-Type":                 "application/json",
			"Access-Control-Allow-Origin":  "*",
			"Access-Control-Allow-Headers": "Content-Type,X-Amz-Date,Authorization,X-Api-Key,X-Amz-Security-Token",
			"Access-Control-Allow-Methods": "GET,OPTIONS",
		},
		Body: string(response),
	}, nil
}

func main() {
	lambda.Start(handler)
}