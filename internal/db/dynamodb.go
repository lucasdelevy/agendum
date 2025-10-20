package db

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

type DynamoDBClient struct {
	Client *dynamodb.Client
}

func NewDynamoDBClient() (*DynamoDBClient, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, err
	}

	return &DynamoDBClient{
		Client: dynamodb.NewFromConfig(cfg),
	}, nil
}