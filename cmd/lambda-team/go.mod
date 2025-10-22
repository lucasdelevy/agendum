module lambda-team

go 1.21

replace agendum => ../..

require (
	agendum v0.0.0-00010101000000-000000000000
	github.com/aws/aws-lambda-go v1.41.0
	github.com/aws/aws-sdk-go v1.45.0
)

require github.com/jmespath/go-jmespath v0.4.0 // indirect
