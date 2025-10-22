# Agendum

## Local Development

```bash
go run cmd/server/main.go
```

## Infrastructure Deployment

### Configure AWS Account
```bash
aws configure --profile target-account
# Enter credentials for target account
export AWS_PROFILE=target-account
```

### Build Lambda Functions
```bash
cd cmd/lambda-user && go mod tidy && make build && cd ..
cd lambda-task && go mod tidy && make build && cd ..
cd lambda-team && go mod tidy && make build && cd ..
cd ../infrastructure
```

### Deploy Infrastructure
```bash
go mod tidy
```

If first time:
```
cdk bootstrap
```

```
cdk deploy --require-approval never
```

## API Endpoints

### Users API
POST `/users/create`
```json
{
  "username": "john_doe",
  "firstName": "John", 
  "lastName": "Doe",
  "userType": "admin"
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
  "owner": "john_doe"
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

## Testing

**Local:**
```bash
curl -X POST http://localhost:8080/users/create/ \
  -H "Content-Type: application/json" \
  -d '{"username":"john_doe","firstName":"John","lastName":"Doe","userType":"admin"}'
```

**AWS (Beta):**
```bash
# Create User
curl -X POST https://ru9oiemsz4.execute-api.us-east-1.amazonaws.com/prod/users/create \
  -H "Content-Type: application/json" \
  -d '{"username":"john_doe","firstName":"John","lastName":"Doe","userType":"admin"}'

# Create Task
curl -X POST https://ru9oiemsz4.execute-api.us-east-1.amazonaws.com/prod/tasks/create \
  -H "Content-Type: application/json" \
  -d '{"title":"Daily Standup","team_id":"team-123","schedule":{"monday":{"begin_time":"09:00","end_time":"09:30"},"tuesday":{"begin_time":"09:00","end_time":"09:30"},"wednesday":{"begin_time":"09:00","end_time":"09:30"},"thursday":{"begin_time":"09:00","end_time":"09:30"},"friday":{"begin_time":"09:00","end_time":"09:30"}},"task_type":"meeting","owner":"john_doe"}'

# Create Team
curl -X POST https://ru9oiemsz4.execute-api.us-east-1.amazonaws.com/prod/teams/create \
  -H "Content-Type: application/json" \
  -d '{"name":"Development Team","admins":["john_doe","jane_smith"],"members":["alice_jones","bob_wilson"]}'
```