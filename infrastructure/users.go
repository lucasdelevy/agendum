package main

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsapigateway"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsdynamodb"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

func NewUserStack(scope constructs.Construct, id string, stage string, props *awscdk.StackProps) awscdk.Stack {
	stack := awscdk.NewStack(scope, &id, props)

	NewUserInfrastructure(stack, "UserInfra", stage)

	return stack
}

func NewUserInfrastructure(scope constructs.Construct, id string, stage string) {
	// DynamoDB Tables
	usersTable := awsdynamodb.NewTable(scope, jsii.String(stage+"-Users"), &awsdynamodb.TableProps{
		TableName: jsii.String(stage + "-Users"),
		PartitionKey: &awsdynamodb.Attribute{
			Name: jsii.String("username"),
			Type: awsdynamodb.AttributeType_STRING,
		},
		RemovalPolicy: awscdk.RemovalPolicy_DESTROY,
	})

	tasksTable := awsdynamodb.NewTable(scope, jsii.String(stage+"-Tasks"), &awsdynamodb.TableProps{
		TableName: jsii.String(stage + "-Tasks"),
		PartitionKey: &awsdynamodb.Attribute{
			Name: jsii.String("task_id"),
			Type: awsdynamodb.AttributeType_STRING,
		},
		RemovalPolicy: awscdk.RemovalPolicy_DESTROY,
	})

	teamsTable := awsdynamodb.NewTable(scope, jsii.String(stage+"-Teams"), &awsdynamodb.TableProps{
		TableName: jsii.String(stage + "-Teams"),
		PartitionKey: &awsdynamodb.Attribute{
			Name: jsii.String("team_id"),
			Type: awsdynamodb.AttributeType_STRING,
		},
		RemovalPolicy: awscdk.RemovalPolicy_DESTROY,
	})

	// Lambda Functions
	createUserLambda := awslambda.NewFunction(scope, jsii.String(stage+"-CreateUserLambda"), &awslambda.FunctionProps{
		FunctionName: jsii.String(stage + "-CreateUserLambda"),
		Runtime: awslambda.Runtime_PROVIDED_AL2023(),
		Handler: jsii.String("bootstrap"),
		Code:    awslambda.Code_FromAsset(jsii.String("../cmd/lambda-user"), nil),
		Environment: &map[string]*string{
			"TABLE_NAME": usersTable.TableName(),
		},
	})

	createTaskLambda := awslambda.NewFunction(scope, jsii.String(stage+"-CreateTaskLambda"), &awslambda.FunctionProps{
		FunctionName: jsii.String(stage + "-CreateTaskLambda"),
		Runtime: awslambda.Runtime_PROVIDED_AL2023(),
		Handler: jsii.String("bootstrap"),
		Code:    awslambda.Code_FromAsset(jsii.String("../cmd/lambda-task"), nil),
		Environment: &map[string]*string{
			"TABLE_NAME": tasksTable.TableName(),
		},
	})

	createTeamLambda := awslambda.NewFunction(scope, jsii.String(stage+"-CreateTeamLambda"), &awslambda.FunctionProps{
		FunctionName: jsii.String(stage + "-CreateTeamLambda"),
		Runtime: awslambda.Runtime_PROVIDED_AL2023(),
		Handler: jsii.String("bootstrap"),
		Code:    awslambda.Code_FromAsset(jsii.String("../cmd/lambda-team"), nil),
		Environment: &map[string]*string{
			"TABLE_NAME": teamsTable.TableName(),
		},
	})

	// Grant permissions
	usersTable.GrantWriteData(createUserLambda)
	tasksTable.GrantWriteData(createTaskLambda)
	teamsTable.GrantWriteData(createTeamLambda)

	// API Gateway
	api := awsapigateway.NewRestApi(scope, jsii.String(stage+"-AgendumApi"), &awsapigateway.RestApiProps{
		RestApiName: jsii.String(stage + "-Agendum API"),
	})

	// Users endpoints
	users := api.Root().AddResource(jsii.String("users"), nil)
	usersCreate := users.AddResource(jsii.String("create"), nil)
	usersCreate.AddMethod(jsii.String("POST"), awsapigateway.NewLambdaIntegration(createUserLambda, nil), nil)

	// Tasks endpoints
	tasks := api.Root().AddResource(jsii.String("tasks"), nil)
	tasksCreate := tasks.AddResource(jsii.String("create"), nil)
	tasksCreate.AddMethod(jsii.String("POST"), awsapigateway.NewLambdaIntegration(createTaskLambda, nil), nil)

	// Teams endpoints
	teams := api.Root().AddResource(jsii.String("teams"), nil)
	teamsCreate := teams.AddResource(jsii.String("create"), nil)
	teamsCreate.AddMethod(jsii.String("POST"), awsapigateway.NewLambdaIntegration(createTeamLambda, nil), nil)
}