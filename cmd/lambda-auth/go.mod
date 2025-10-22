module lambda-auth

go 1.21

replace agendum => ../..

require (
	github.com/aws/aws-lambda-go v1.41.0
	github.com/aws/aws-sdk-go v1.45.0
	golang.org/x/crypto v0.17.0
)

require github.com/jmespath/go-jmespath v0.4.0 // indirect
