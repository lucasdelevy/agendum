module lambda-list-teams

go 1.21

require (
	agendum v0.0.0-00010101000000-000000000000
	github.com/aws/aws-lambda-go v1.47.0
	github.com/aws/aws-sdk-go v1.55.5
)

replace agendum => ../../

require github.com/jmespath/go-jmespath v0.4.0 // indirect
