# Agendum

## Usage

```bash
go run cmd/server/main.go
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

```bash
curl -X POST http://localhost:8080/users/create/ \
  -H "Content-Type: application/json" \
  -d '{"username":"john_doe","firstName":"John","lastName":"Doe","userType":"admin"}'
```