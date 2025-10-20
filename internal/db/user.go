package db

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"agendum/internal/models"
)

func (d *DynamoDBClient) CreateUser(ctx context.Context, user models.User) error {
	_, err := d.Client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String("beta-Users"),
		Item: map[string]types.AttributeValue{
			"username":  &types.AttributeValueMemberS{Value: user.Username},
			"firstName": &types.AttributeValueMemberS{Value: user.FirstName},
			"lastName":  &types.AttributeValueMemberS{Value: user.LastName},
			"userType":  &types.AttributeValueMemberS{Value: user.UserType},
		},
	})
	return err
}