# User Management Service

A comprehensive user management backend service built using the Lokstra framework with DSL-based flow handling.

## Features

- **Complete CRUD Operations**: Create, Read, Update, Delete users
- **Multi-tenant Support**: Tenant-based user isolation
- **DSL-based Request Handling**: Uses Lokstra's DSL for robust flow processing
- **Metrics Integration**: Comprehensive metrics collection with Prometheus
- **I18n Support**: Multi-language error messages and localization
- **Audit Logging**: Complete audit trail for all user operations
- **Graceful Shutdown**: Proper resource cleanup and connection management
- **Database Migrations**: SQL schema with indexes and constraints

## Architecture

The service follows a clean architecture pattern:

```
cmd/projects/user_management/
├── main.go                     # Application entry point
├── config/
│   └── user_management.yaml    # Service configuration
├── internal/
│   ├── handlers/              # HTTP request handlers using DSL
│   ├── repository/            # Data access layer implementing auth.UserRepository
│   └── models/               # Request/response models and validation
└── migrations/               # Database schema migrations
```

### Key Components

1. **DSL-Enhanced Repository**: Implements `serviceapi/auth.UserRepository` using DSL flows
2. **Flow-based Handlers**: HTTP handlers that use DSL for request processing
3. **Enhanced Error Handling**: Localized errors with I18n support
4. **Comprehensive Metrics**: Performance and business metrics collection

## API Endpoints

All endpoints support multi-tenant operations via the `X-Tenant-ID` header.

### Create User
```http
POST /api/v1/users
Content-Type: application/json
X-Tenant-ID: your-tenant-id

{
  "username": "johndoe",
  "email": "john@example.com", 
  "password": "securepassword123",
  "is_active": true,
  "metadata": {
    "department": "engineering",
    "role": "developer"
  }
}
```

### Get User
```http
GET /api/v1/user?username=johndoe
X-Tenant-ID: your-tenant-id
```

### Update User
```http
PUT /api/v1/user?username=johndoe
Content-Type: application/json
X-Tenant-ID: your-tenant-id

{
  "email": "newemail@example.com",
  "is_active": false,
  "metadata": {
    "department": "marketing"
  }
}
```

### Delete User (Soft Delete)
```http
DELETE /api/v1/user?username=johndoe
X-Tenant-ID: your-tenant-id
```

### List Users
```http
GET /api/v1/users?limit=10&offset=0
X-Tenant-ID: your-tenant-id
```

### Health Check
```http
GET /api/v1/health
```

## Configuration

The service is configured via YAML file (`config/user_management.yaml`):

```yaml
services:
  db_pool:
    type: "dbpool_pg"
    config:
      database_url: "postgres://user:password@localhost/userdb"
      max_connections: 10
      min_connections: 2

  logger:
    type: "logger"
    config:
      level: "info"
      format: "json"

  metrics:
    type: "metrics"
    config:
      enabled: true
      namespace: "user_management"

  i18n:
    type: "i18n"
    config:
      default_language: "en"
      supported_languages: ["en", "id", "es"]

modules:
  http_listener:
    bind_address: "0.0.0.0:8080"
    read_timeout: "15s"
    write_timeout: "15s"

  http_router:
    cors_enabled: true
    request_logging: true
```

## Database Schema

The service uses PostgreSQL with the following main tables:

- **users**: Main user information with soft delete support
- **user_sessions**: Session management (optional)
- **user_permissions**: Role-based access control (optional)
- **user_audit_log**: Complete audit trail

### Key Features of Schema:
- Unique constraints on username/email per tenant
- Partial indexes for active users only
- JSONB metadata support
- Automatic timestamp updates
- Comprehensive indexing for performance

## Running the Service

### Prerequisites
- Go 1.24.4+
- PostgreSQL 12+
- Lokstra framework

### Database Setup
1. Create PostgreSQL database
2. Run migration:
```bash
psql -d userdb -f migrations/001_create_users_tables.sql
```

### Starting the Service
```bash
# Using default config
go run main.go

# Using custom config  
go run main.go path/to/config.yaml
```

### Environment Variables
- `DB_URL`: Database connection string
- `LOG_LEVEL`: Logging level (debug, info, warn, error)
- `PORT`: HTTP server port (default: 8080)

## DSL Flow Examples

### User Creation Flow
```go
flow := dsl.NewFlow("CreateUser", serviceVars)

flow.Validate(validateInput).
     ErrorIfExists(usernameError, checkUsernameSQL, args...).
     ErrorIfExists(emailError, checkEmailSQL, args...).
     BeginTx().
     ExecSql(insertUserSQL, args...).
     CommitOrRollback()
```

### Query with Metrics
```go
flow.QueryOneSaveAs(selectUserSQL, "user_row", tenantID, username).
     Do(convertRowToUser).
     // Automatic metrics collection for query time, success/failure
```

## Error Handling

The service provides localized error messages:

```json
{
  "error": "Username is required",
  "status": "error", 
  "code": 400
}
```

Supported error codes:
- `validation.required_field`
- `validation.invalid_value`
- `resource.not_found`
- `database.operation_failed`

## Metrics

Comprehensive metrics are collected:

### Business Metrics
- `user_operations_total{operation, tenant_id, status}`
- `users_created_total{tenant_id}`
- `users_active_total{tenant_id}`

### Performance Metrics  
- `http_requests_duration_seconds{method, endpoint}`
- `database_queries_duration_seconds{operation}`
- `flow_execution_duration_seconds{flow_name}`

### Error Metrics
- `errors_total{type, code}`
- `validation_failures_total{field}`

## Testing

### Unit Tests
```bash
go test ./internal/...
```

### Integration Tests
```bash
# Requires running PostgreSQL
go test ./... -tags=integration
```

### Manual Testing
```bash
# Create user
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: test" \
  -d '{"username":"test","email":"test@example.com","password":"password123"}'

# Get user  
curl http://localhost:8080/api/v1/user?username=test \
  -H "X-Tenant-ID: test"
```

## Development

### Adding New Endpoints
1. Add method to `handlers/user_handler.go`
2. Create DSL flow for business logic
3. Add route in `main.go`
4. Update tests

### Adding New Validations
1. Add validation logic to `models/user.go`
2. Create localized error messages
3. Add test cases

### Extending DSL
1. Add new step types in `core/dsl/`
2. Add helper methods to Flow
3. Update documentation

## Production Considerations

### Security
- Use bcrypt for password hashing (current implementation uses SHA256 for demo)
- Implement rate limiting
- Add input sanitization
- Use HTTPS in production

### Performance
- Enable database connection pooling
- Add Redis caching layer
- Implement pagination for large datasets
- Use read replicas for queries

### Monitoring
- Set up Prometheus metrics collection
- Configure alerting rules
- Add health check endpoints
- Implement distributed tracing

### Deployment
- Use containerization (Docker)
- Set up load balancing
- Configure auto-scaling
- Implement blue-green deployments

## Contributing

1. Fork the repository
2. Create feature branch
3. Add tests for new functionality
4. Ensure all tests pass
5. Submit pull request

## License

This project is part of the Lokstra framework and follows its licensing terms.
