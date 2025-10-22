package auth

import (
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

// ValidateToken checks if a token is valid and returns the username
func ValidateToken(token string) (string, bool) {
	sess := session.Must(session.NewSession())
	svc := dynamodb.New(sess)

	result, err := svc.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String(os.Getenv("SESSIONS_TABLE_NAME")),
		Key: map[string]*dynamodb.AttributeValue{
			"token": {S: aws.String(token)},
		},
	})

	if err != nil || result.Item == nil {
		return "", false
	}

	// Check if token is expired
	expiresAt := *result.Item["expires_at"].S
	expireTime, err := time.Parse(time.RFC3339, expiresAt)
	if err != nil || time.Now().After(expireTime) {
		return "", false
	}

	username := *result.Item["username"].S
	return username, true
}

// IsTeamAdmin checks if a user is an admin of a specific team
func IsTeamAdmin(username, teamID string) bool {
	sess := session.Must(session.NewSession())
	svc := dynamodb.New(sess)

	result, err := svc.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String(os.Getenv("TEAMS_TABLE_NAME")),
		Key: map[string]*dynamodb.AttributeValue{
			"team_id": {S: aws.String(teamID)},
		},
	})

	if err != nil || result.Item == nil {
		return false
	}

	adminsStr := *result.Item["admins"].S
	admins := strings.Split(adminsStr, ",")
	
	for _, admin := range admins {
		if strings.TrimSpace(admin) == username {
			return true
		}
	}
	
	return false
}