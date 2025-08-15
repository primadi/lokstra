# Best Practice: Application Architecture

This example demonstrates best practices for structuring a Lokstra application with reusable modules, proper separation of concerns, and database integration.

## Project Structure

```
02_application_architecture/
├── README.md                    # This documentation
├── lokstra.yaml                 # Main configuration file
├── main.go                      # Application entry point
├── go.mod                       # Go module definition
├── modules/                     # Reusable modules directory
│   └── user_management/         # User management module
│       ├── module.go            # Module registration
│       ├── handlers/            # HTTP handlers
│       │   └── user_handler.go
│       ├── repository/          # Data access layer
│       │   └── user_repository.go
│       ├── models/              # Data models
│       │   └── user.go
│       └── services/            # Business logic layer
│           └── user_service.go
└── migrations/                  # Database migrations
    └── 001_create_users_table.sql
```

## Features Demonstrated

### 1. Module Architecture
- **Reusable Modules**: The `user_management` module is designed to be portable across projects
- **Separation of Concerns**: Clear separation between handlers, services, repositories, and models
- **Interface-Based Design**: Uses interfaces for testability and flexibility

### 2. Database Integration
- **Connection Pooling**: Uses `dbpool_pg` service for efficient database connections
- **Repository Pattern**: Clean data access layer with interface abstractions
- **Transaction Support**: Proper transaction handling for data consistency

### 3. RESTful API Design
- **CRUD Operations**: Complete Create, Read, Update, Delete operations
- **List with Pagination**: Efficient list operations with pagination support
- **Proper HTTP Status Codes**: Correct status codes for different operations
- **Error Handling**: Consistent error responses

### 4. Configuration Management
- **YAML Configuration**: Uses lokstra.yaml for easy configuration management
- **Environment-Specific Settings**: Supports different configurations per environment
- **Service Dependencies**: Proper service dependency management

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET    | /api/users | List all users (with pagination) |
| GET    | /api/users/:id | Get user by ID |
| POST   | /api/users | Create new user |
| PUT    | /api/users/:id | Update user |
| DELETE | /api/users/:id | Delete user |

## Quick Start

1. **Install Dependencies**:
   ```bash
   go mod tidy
   ```

2. **Setup Database**:
   ```bash
   # Create PostgreSQL database
   createdb lokstra_example
   
   # Run migrations
   psql lokstra_example < migrations/001_create_users_table.sql
   ```

3. **Update Configuration**:
   Edit `lokstra.yaml` and update the database connection string.

4. **Run the Application**:
   ```bash
   go run main.go
   ```

5. **Test the API**:
   ```bash
   # Create a user
   curl -X POST http://localhost:8080/api/users \
     -H "Content-Type: application/json" \
     -d '{"name":"John Doe","email":"john@example.com"}'
   
   # List users
   curl http://localhost:8080/api/users
   
   # Get user by ID
   curl http://localhost:8080/api/users/1
   ```

## Best Practices Implemented

### 1. **Module Design**
- Self-contained modules with clear boundaries
- Interface-driven development for testability
- Dependency injection for loose coupling

### 2. **Error Handling**
- Consistent error response format
- Proper HTTP status codes
- Detailed error messages for debugging

### 3. **Database Design**
- Connection pooling for performance
- Repository pattern for data access abstraction
- Transaction support for data integrity

### 4. **Configuration Management**
- Centralized configuration in YAML
- Environment-specific overrides
- Service dependency declarations

### 5. **Code Organization**
- Clear separation of concerns
- Consistent naming conventions
- Proper package structure

## Extending This Example

This module can be extended by:

1. **Adding Authentication**: Integrate with JWT auth module
2. **Adding Validation**: Implement request validation
3. **Adding Caching**: Add Redis caching layer
4. **Adding Tests**: Comprehensive unit and integration tests
5. **Adding Logging**: Structured logging throughout the application

## Module Reusability

The `user_management` module is designed to be reusable:

1. **Copy the Module**: Copy `modules/user_management` to your project
2. **Update Configuration**: Add the module to your lokstra.yaml
3. **Register Routes**: Mount the module routes in your application
4. **Customize**: Modify handlers and services as needed

This demonstrates how Lokstra modules can be shared across different projects while maintaining clean architecture and separation of concerns.
