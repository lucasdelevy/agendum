package main

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsapigateway"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsdynamodb"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

func NewAgendumInfrastructure(scope constructs.Construct, id string) {
	// DynamoDB Table
	table := awsdynamodb.NewTable(scope, jsii.String("beta-Users"), &awsdynamodb.TableProps{
		TableName: jsii.String("beta-Users"),
		PartitionKey: &awsdynamodb.Attribute{
			Name: jsii.String("username"),
			Type: awsdynamodb.AttributeType_STRING,
		},
		RemovalPolicy: awscdk.RemovalPolicy_DESTROY,
	})

	// Lambda Function
	lambda := awslambda.NewFunction(scope, jsii.String("beta-CreateUserLambda"), &awslambda.FunctionProps{
		FunctionName: jsii.String("beta-CreateUserLambda"),
		Runtime: awslambda.Runtime_PROVIDED_AL2023(),
		Handler: jsii.String("bootstrap"),
		Code:    awslambda.Code_FromAsset(jsii.String("../cmd/lambda"), nil),
		Environment: &map[string]*string{
			"TABLE_NAME": table.TableName(),
		},
	})

	table.GrantWriteData(lambda)

	// API Gateway
	api := awsapigateway.NewRestApi(scope, jsii.String("beta-AgendumApi"), &awsapigateway.RestApiProps{
		RestApiName: jsii.String("beta-Agendum API"),
	})

	users := api.Root().AddResource(jsii.String("users"), nil)
	create := users.AddResource(jsii.String("create"), nil)

	create.AddMethod(jsii.String("POST"), awsapigateway.NewLambdaIntegration(lambda, nil), nil)
}