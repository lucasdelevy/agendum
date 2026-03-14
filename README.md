# Agendum

A task and team management API built with Go, deployed to AWS using Lambda, API Gateway, and DynamoDB.

## Prerequisites

- [Go 1.21+](https://go.dev/dl/)
- [AWS CLI](https://aws.amazon.com/cli/) (configured with credentials)
- [AWS CDK](https://docs.aws.amazon.com/cdk/v2/guide/getting-started.html) (`npm install -g aws-cdk`)
- [Node.js 18+](https://nodejs.org/) (required by CDK)

## Local Development

```bash
make run
```

This starts the local HTTP server on port 8080.

## Deployment

### Quick Deploy (one command)

```bash
./deploy.sh
```

For first-time setup (bootstraps CDK):

```bash
./deploy.sh --bootstrap
```

### Using Make

```bash
# Build all Lambda functions
make build

# Build and deploy to AWS
make deploy

# First-time CDK bootstrap
make bootstrap

# Clean build artifacts
make clean
```

### Manual Step-by-Step

1. Configure AWS credentials:
```bash
aws configure --profile target-account
export AWS_PROFILE=target-account
```

2. Build all Lambda functions:
```bash
make build
```

3. Bootstrap CDK (first time only):
```bash
make bootstrap
```

4. Deploy:
```bash
make deploy
```

## CI/CD

This project includes a GitHub Actions workflow (`.github/workflows/deploy.yml`) that:

- **On pull requests**: Builds all Lambda functions to validate the code compiles
- **On push to main**: Builds and deploys to AWS automatically

To enable CI/CD, add these secrets to your GitHub repository:

| Secret | Description |
|--------|-------------|
| `AWS_ROLE_ARN` | IAM role ARN for GitHub Actions OIDC authentication |
| `AWS_REGION` | AWS region to deploy to (e.g., `us-east-1`) |

## API Endpoints

### Users API
POST `/users/create`
```json
{
  "username": "john_doe",
  "email": "john@example.com",
  "password": "password123",
  "firstName": "John", 
  "lastName": "Doe",
  "userType": "admin"
}
```

### Auth API
POST `/auth/login`
```json
{
  "email": "john@example.com",
  "password": "password123"
}
```

### Tasks API
POST `/tasks/create`
```json
{
  "title": "Daily Standup",
  "team_id": "team-123",
  "schedule": {
    "monday": {
      "begin_time": "09:00",
      "end_time": "09:30"
    },
    "tuesday": {
      "begin_time": "09:00",
      "end_time": "09:30"
    },
    "wednesday": {
      "begin_time": "09:00",
      "end_time": "09:30"
    },
    "thursday": {
      "begin_time": "09:00",
      "end_time": "09:30"
    },
    "friday": {
      "begin_time": "09:00",
      "end_time": "09:30"
    }
  },
  "task_type": "meeting",
  "requester": "john_doe"
}
```

**Note:** Times are in HH:MM format. Only include days when the task occurs.

### Teams API
POST `/teams/create`
```json
{
  "name": "Development Team",
  "admins": ["john_doe", "jane_smith"],
  "members": ["alice_jones", "bob_wilson"]
}
```

GET `/teams/list` (requires auth token)
Returns list of teams the authenticated user belongs to.

Response:
```json
[
  {
    "team_id": "team-123",
    "name": "Development Team",
    "admins": ["john_doe", "jane_smith"],
    "members": ["alice_jones", "bob_wilson"]
  }
]
```

## Testing

**Local:**
```bash
curl -X POST http://localhost:8080/users/create/ \
  -H "Content-Type: application/json" \
  -d '{"username":"john_doe","firstName":"John","lastName":"Doe","userType":"admin"}'
```

**AWS (Beta):**

### Authentication Workflow

**Step 1: Create User**
```bash
curl -X POST https://u7zrjhuptb.execute-api.us-east-1.amazonaws.com/prod/users/create \
  -H "Content-Type: application/json" \
  -d '{"username":"john_doe","email":"john@example.com","password":"password123","firstName":"John","lastName":"Doe","userType":"admin"}'
```

**Step 2: Login to get token**
```bash
curl -X POST https://u7zrjhuptb.execute-api.us-east-1.amazonaws.com/prod/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"john@example.com","password":"password123"}'
```

**Step 3: Create Team (requires auth token)**
```bash
curl -X POST https://u7zrjhuptb.execute-api.us-east-1.amazonaws.com/prod/teams/create \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN_FROM_STEP_2" \
  -d '{"name":"Development Team","admins":["john_doe","jane_smith"],"members":["alice_jones","bob_wilson"]}'
```

**Step 4: List Teams (requires auth token)**
```bash
curl -X GET https://u7zrjhuptb.execute-api.us-east-1.amazonaws.com/prod/teams/list \
  -H "Authorization: Bearer YOUR_TOKEN_FROM_STEP_2"
```

**Step 5: Create Task (requires auth token + team admin)**
```bash
curl -X POST https://u7zrjhuptb.execute-api.us-east-1.amazonaws.com/prod/tasks/create \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN_FROM_STEP_2" \
  -d '{"title":"Daily Standup","team_id":"TEAM_ID_FROM_STEP_3","schedule":{"monday":{"begin_time":"09:00","end_time":"09:30"},"tuesday":{"begin_time":"09:00","end_time":"09:30"},"wednesday":{"begin_time":"09:00","end_time":"09:30"},"thursday":{"begin_time":"09:00","end_time":"09:30"},"friday":{"begin_time":"09:00","end_time":"09:30"}},"task_type":"meeting"}'
```

**Note:** 
- Replace `YOUR_TOKEN_FROM_STEP_2` with the actual token returned from login
- Replace `TEAM_ID_FROM_STEP_3` with the team_id returned from team creation
- Tasks can only be created by team admins
- Tokens expire after 24 hours