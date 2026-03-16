LAMBDA_DIRS := cmd/lambda-user cmd/lambda-auth cmd/lambda-task cmd/lambda-team cmd/lambda-list-teams
INFRASTRUCTURE_DIR := infrastructure

.PHONY: all build clean deploy bootstrap tidy test run

all: build

build: tidy
	@echo "Building all Lambda functions..."
	@for dir in $(LAMBDA_DIRS); do \
		echo "  Building $$dir..."; \
		cd $$dir && GOOS=linux GOARCH=amd64 go build -o bootstrap main.go && cd $(CURDIR); \
	done
	@echo "All Lambda functions built successfully."

tidy:
	@echo "Tidying Go modules..."
	@for dir in $(LAMBDA_DIRS); do \
		cd $$dir && go mod tidy && cd $(CURDIR); \
	done
	@cd $(INFRASTRUCTURE_DIR) && go mod tidy
	@echo "All modules tidied."

clean:
	@echo "Cleaning build artifacts..."
	@for dir in $(LAMBDA_DIRS); do \
		rm -f $$dir/bootstrap; \
	done
	@rm -rf $(INFRASTRUCTURE_DIR)/cdk.out
	@echo "Clean complete."

bootstrap:
	@echo "Bootstrapping CDK (first-time setup)..."
	cd $(INFRASTRUCTURE_DIR) && cdk bootstrap

deploy: build
	@echo "Deploying infrastructure with CDK..."
	cd $(INFRASTRUCTURE_DIR) && cdk deploy --require-approval never
	@echo "Deployment complete."

test:
	@echo "Running tests..."
	go test ./...
	@for dir in $(LAMBDA_DIRS); do \
		cd $$dir && go test ./... && cd $(CURDIR); \
	done

run:
	go run cmd/server/main.go
