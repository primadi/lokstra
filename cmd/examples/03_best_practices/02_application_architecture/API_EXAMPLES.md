# API Testing Examples

This file contains examples for testing the User Management API endpoints using curl commands.

## Prerequisites

1. Make sure the application is running:
   ```bash
   go run main.go
   ```

2. Ensure the database is set up with the migration:
   ```bash
   psql your_database < migrations/001_create_users_table.sql
   ```

## API Examples

### 1. Health Check
```bash
curl -X GET http://localhost:8080/api/health
```

Expected Response:
```json
{
  "status": "healthy",
  "service": "application_architecture_example",
  "version": "1.0.0",
  "description": "Lokstra Application Architecture Best Practices Example"
}
```

### 2. Create a User
```bash
curl -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Alice Johnson",
    "email": "alice.johnson@example.com"
  }'
```

Expected Response:
```json
{
  "user": {
    "id": 4,
    "name": "Alice Johnson",
    "email": "alice.johnson@example.com",
    "created_at": "2024-01-15T10:30:00Z",
    "updated_at": "2024-01-15T10:30:00Z"
  }
}
```

### 3. List All Users (with pagination)
```bash
# Default: page=1, page_size=10
curl -X GET http://localhost:8080/api/users

# With pagination parameters
curl -X GET "http://localhost:8080/api/users?page=1&page_size=5"

# With search
curl -X GET "http://localhost:8080/api/users?search=john"
```

Expected Response:
```json
{
  "users": [
    {
      "id": 1,
      "name": "John Doe",
      "email": "john.doe@example.com",
      "created_at": "2024-01-15T09:00:00Z",
      "updated_at": "2024-01-15T09:00:00Z"
    },
    {
      "id": 2,
      "name": "Jane Smith",
      "email": "jane.smith@example.com",
      "created_at": "2024-01-15T09:15:00Z",
      "updated_at": "2024-01-15T09:15:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "page_size": 10,
    "total_items": 4,
    "total_pages": 1
  }
}
```

### 4. Get User by ID
```bash
curl -X GET http://localhost:8080/api/users/1
```

Expected Response:
```json
{
  "user": {
    "id": 1,
    "name": "John Doe",
    "email": "john.doe@example.com",
    "created_at": "2024-01-15T09:00:00Z",
    "updated_at": "2024-01-15T09:00:00Z"
  }
}
```

### 5. Update User
```bash
# Update name only
curl -X PUT http://localhost:8080/api/users/1 \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Doe Updated"
  }'

# Update email only
curl -X PUT http://localhost:8080/api/users/1 \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john.doe.updated@example.com"
  }'

# Update both name and email
curl -X PUT http://localhost:8080/api/users/1 \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Doe Complete Update",
    "email": "john.complete@example.com"
  }'
```

Expected Response:
```json
{
  "user": {
    "id": 1,
    "name": "John Doe Updated",
    "email": "john.complete@example.com",
    "created_at": "2024-01-15T09:00:00Z",
    "updated_at": "2024-01-15T10:45:00Z"
  }
}
```

### 6. Delete User
```bash
curl -X DELETE http://localhost:8080/api/users/1
```

Expected Response:
```json
{
  "message": "User deleted successfully"
}
```

## Error Scenarios

### 1. Create User with Duplicate Email
```bash
curl -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Duplicate User",
    "email": "john.doe@example.com"
  }'
```

Expected Response (409 Conflict):
```json
{
  "error": "User with this email already exists"
}
```

### 2. Get Non-existent User
```bash
curl -X GET http://localhost:8080/api/users/999
```

Expected Response (404 Not Found):
```json
{
  "error": "User not found"
}
```

### 3. Create User with Invalid Data
```bash
curl -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -d '{
    "name": "",
    "email": "invalid-email"
  }'
```

Expected Response (400 Bad Request):
```json
{
  "error": "Validation failed",
  "field": "name",
  "message": "Name is required"
}
```

### 4. Invalid JSON
```bash
curl -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -d '{"name": "Test", "email":}'
```

Expected Response (400 Bad Request):
```json
{
  "error": "Invalid JSON in request body"
}
```

## Batch Testing Script

Create a file named `test_api.sh`:

```bash
#!/bin/bash

BASE_URL="http://localhost:8080/api"

echo "🧪 Testing User Management API..."

echo "1. Testing health check..."
curl -s "$BASE_URL/health" | jq .

echo -e "\n2. Creating test user..."
USER_RESPONSE=$(curl -s -X POST "$BASE_URL/users" \
  -H "Content-Type: application/json" \
  -d '{"name":"Test User","email":"test@example.com"}')
echo $USER_RESPONSE | jq .

USER_ID=$(echo $USER_RESPONSE | jq -r '.user.id')

echo -e "\n3. Getting user by ID ($USER_ID)..."
curl -s "$BASE_URL/users/$USER_ID" | jq .

echo -e "\n4. Updating user..."
curl -s -X PUT "$BASE_URL/users/$USER_ID" \
  -H "Content-Type: application/json" \
  -d '{"name":"Updated Test User"}' | jq .

echo -e "\n5. Listing all users..."
curl -s "$BASE_URL/users" | jq .

echo -e "\n6. Deleting user..."
curl -s -X DELETE "$BASE_URL/users/$USER_ID" | jq .

echo -e "\n✅ API testing completed!"
```

Make it executable and run:
```bash
chmod +x test_api.sh
./test_api.sh
```
