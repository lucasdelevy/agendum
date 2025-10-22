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

type TimeSlot struct {
	BeginTime string `json:"begin_time"`
	EndTime   string `json:"end_time"`
}

type WeeklySchedule struct {
	Monday    *TimeSlot `json:"monday,omitempty"`
	Tuesday   *TimeSlot `json:"tuesday,omitempty"`
	Wednesday *TimeSlot `json:"wednesday,omitempty"`
	Thursday  *TimeSlot `json:"thursday,omitempty"`
	Friday    *TimeSlot `json:"friday,omitempty"`
	Saturday  *TimeSlot `json:"saturday,omitempty"`
	Sunday    *TimeSlot `json:"sunday,omitempty"`
}

type Task struct {
	Title     string          `json:"title"`
	TeamID    string          `json:"team_id"`
	Schedule  *WeeklySchedule `json:"schedule"`
	TaskType  string          `json:"task_type"`
	Requester string          `json:"requester"`
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

	username, valid := auth.ValidateToken(token)
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

	var task Task
	if err := json.Unmarshal([]byte(request.Body), &task); err != nil {
		return events.APIGatewayProxyResponse{StatusCode: 400}, err
	}

	// Set requester to authenticated username
	task.Requester = username

	// Check if user is admin of the team
	if !auth.IsTeamAdmin(username, task.TeamID) {
		return events.APIGatewayProxyResponse{
			StatusCode: 403,
			Headers: map[string]string{
				"Content-Type":                 "application/json",
				"Access-Control-Allow-Origin":  "*",
				"Access-Control-Allow-Headers": "Content-Type,X-Amz-Date,Authorization,X-Api-Key,X-Amz-Security-Token",
				"Access-Control-Allow-Methods": "POST,OPTIONS",
			},
			Body: `{"message":"Only team admins can create tasks"}`,
		}, nil
	}

	sess := session.Must(session.NewSession())
	svc := dynamodb.New(sess)

	taskID := utils.GenerateID()
	createdTimestamp := time.Now().Format(time.RFC3339)

	scheduleMap := make(map[string]*dynamodb.AttributeValue)
	if task.Schedule.Monday != nil {
		scheduleMap["monday"] = &dynamodb.AttributeValue{M: map[string]*dynamodb.AttributeValue{
			"begin_time": {S: aws.String(task.Schedule.Monday.BeginTime)},
			"end_time":   {S: aws.String(task.Schedule.Monday.EndTime)},
		}}
	}
	if task.Schedule.Tuesday != nil {
		scheduleMap["tuesday"] = &dynamodb.AttributeValue{M: map[string]*dynamodb.AttributeValue{
			"begin_time": {S: aws.String(task.Schedule.Tuesday.BeginTime)},
			"end_time":   {S: aws.String(task.Schedule.Tuesday.EndTime)},
		}}
	}
	if task.Schedule.Wednesday != nil {
		scheduleMap["wednesday"] = &dynamodb.AttributeValue{M: map[string]*dynamodb.AttributeValue{
			"begin_time": {S: aws.String(task.Schedule.Wednesday.BeginTime)},
			"end_time":   {S: aws.String(task.Schedule.Wednesday.EndTime)},
		}}
	}
	if task.Schedule.Thursday != nil {
		scheduleMap["thursday"] = &dynamodb.AttributeValue{M: map[string]*dynamodb.AttributeValue{
			"begin_time": {S: aws.String(task.Schedule.Thursday.BeginTime)},
			"end_time":   {S: aws.String(task.Schedule.Thursday.EndTime)},
		}}
	}
	if task.Schedule.Friday != nil {
		scheduleMap["friday"] = &dynamodb.AttributeValue{M: map[string]*dynamodb.AttributeValue{
			"begin_time": {S: aws.String(task.Schedule.Friday.BeginTime)},
			"end_time":   {S: aws.String(task.Schedule.Friday.EndTime)},
		}}
	}
	if task.Schedule.Saturday != nil {
		scheduleMap["saturday"] = &dynamodb.AttributeValue{M: map[string]*dynamodb.AttributeValue{
			"begin_time": {S: aws.String(task.Schedule.Saturday.BeginTime)},
			"end_time":   {S: aws.String(task.Schedule.Saturday.EndTime)},
		}}
	}
	if task.Schedule.Sunday != nil {
		scheduleMap["sunday"] = &dynamodb.AttributeValue{M: map[string]*dynamodb.AttributeValue{
			"begin_time": {S: aws.String(task.Schedule.Sunday.BeginTime)},
			"end_time":   {S: aws.String(task.Schedule.Sunday.EndTime)},
		}}
	}

	item := map[string]*dynamodb.AttributeValue{
		"task_id":           {S: aws.String(taskID)},
		"title":             {S: aws.String(task.Title)},
		"team_id":           {S: aws.String(task.TeamID)},
		"created_timestamp": {S: aws.String(createdTimestamp)},
		"schedule":          {M: scheduleMap},
		"task_type":         {S: aws.String(task.TaskType)},
		"requester":         {S: aws.String(task.Requester)},
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
		Headers: map[string]string{
			"Content-Type":                 "application/json",
			"Access-Control-Allow-Origin":  "*",
			"Access-Control-Allow-Headers": "Content-Type,X-Amz-Date,Authorization,X-Api-Key,X-Amz-Security-Token",
			"Access-Control-Allow-Methods": "POST,OPTIONS",
		},
		Body: `{"message":"Task created successfully","task_id":"` + taskID + `"}`,
	}, nil
}

func main() {
	lambda.Start(handler)
}
