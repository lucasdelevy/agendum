# Agendum

## Local Development

```bash
go run cmd/server/main.go
```

## Infrastructure Deployment

```bash
cd infrastructure
go mod tidy
```

If first time:
```
cdk bootstrap
```

```
cdk deploy --require-approval never
```

## Endpoint

POST `/users/create/`

```json
{
  "username": "john_doe",
  "firstName": "John", 
  "lastName": "Doe",
  "userType": "admin"
}
```

## Testing

**Local:**
```bash
curl -X POST http://localhost:8080/users/create/ \
  -H "Content-Type: application/json" \
  -d '{"username":"john_doe","firstName":"John","lastName":"Doe","userType":"admin"}'
```

**AWS (Deployed):**
```bash
curl -X POST https://0pkc1a4wqk.execute-api.us-east-1.amazonaws.com/prod/users/create \
  -H "Content-Type: application/json" \
  -d '{"username":"john_doe","firstName":"John","lastName":"Doe","userType":"admin"}'
```