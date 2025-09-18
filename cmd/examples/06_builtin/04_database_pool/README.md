# Database Pool Service Example

This example demonstrates how to use Lokstra's built-in PostgreSQL database pool service for database operations.

## Overview

The database pool service provides:
- Connection pooling for PostgreSQL databases
- Automatic connection management
- Health monitoring
- Transaction support
- Connection lifecycle management

## Prerequisites

1. PostgreSQL server running locally or remotely
2. Database accessible at the configured connection string
3. Go modules initialized

## Configuration

The database pool is configured with:

```go
dbConfig := map[string]any{
    "connection_string": "postgres://localhost:5432/testdb?sslmode=disable",
    "max_connections":   10,
    "min_connections":   2,
    "max_idle_time":     "30m",
    "max_lifetime":      "1h",
}
```

### Configuration Options

- `connection_string`: PostgreSQL connection URL
- `max_connections`: Maximum number of connections in pool
- `min_connections`: Minimum number of connections to maintain
- `max_idle_time`: Maximum time a connection can be idle
- `max_lifetime`: Maximum lifetime of a connection

## Running the Example

1. **Start PostgreSQL server**:
   ```bash
   # Using Docker
   docker run -d \
     --name postgres-demo \
     -e POSTGRES_DB=testdb \
     -e POSTGRES_USER=demo \
     -e POSTGRES_PASSWORD=demo \
     -p 5432:5432 \
     postgres:15

   # Or use existing PostgreSQL installation
   ```

2. **Update connection string** in `main.go` if needed

3. **Run the application**:
   ```bash
   go run main.go
   ```

4. **Test the endpoints**:
   ```bash
   # Health check with database status
   curl http://localhost:8080/health

   # Get all users
   curl http://localhost:8080/users

   # Create a new user
   curl -X POST http://localhost:8080/users \
     -H "Content-Type: application/json" \
     -d '{"name":"John Doe","email":"john@example.com"}'

   # Get specific user
   curl http://localhost:8080/users/123

   # Update user
   curl -X PUT http://localhost:8080/users/123 \
     -H "Content-Type: application/json" \
     -d '{"name":"Jane Doe","email":"jane@example.com"}'

   # Delete user
   curl -X DELETE http://localhost:8080/users/123

   # Database statistics
   curl http://localhost:8080/db/stats
   ```

## API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/health` | Health check with database connectivity |
| GET | `/users` | List all users |
| POST | `/users` | Create new user |
| GET | `/users/:id` | Get user by ID |
| PUT | `/users/:id` | Update user |
| DELETE | `/users/:id` | Delete user |
| GET | `/db/stats` | Database connection pool statistics |

## Key Features Demonstrated

### 1. **Service Registration**
```go
err := app.RegistrationContext().RegisterServiceWithConfig("dbpool", dbpool_pg.NewService, dbConfig)
```

### 2. **Connection Acquisition**
```go
db := dbService.(serviceapi.DbPool)
conn, err := db.Acquire(ctx.Request.Context(), "")
defer conn.Release()
```

### 3. **Health Monitoring**
```go
err = conn.Ping(ctx.Request.Context())
```

### 4. **Error Handling**
- Graceful degradation when database is unavailable
- Proper connection cleanup
- Meaningful error messages

## Production Considerations

1. **Connection String Security**:
   - Use environment variables for credentials
   - Consider using connection string builders
   - Enable SSL in production

2. **Pool Configuration**:
   - Tune pool size based on load
   - Monitor connection usage
   - Set appropriate timeouts

3. **Error Handling**:
   - Implement proper transaction rollback
   - Handle connection pool exhaustion
   - Add circuit breaker patterns

4. **Monitoring**:
   - Log slow queries
   - Monitor pool statistics
   - Set up alerts for connection issues

## Real-World Extensions

For production use, consider adding:

- **Schema Migrations**: Database versioning and migration tools
- **Query Builder**: SQL query construction helpers  
- **ORM Integration**: Object-relational mapping
- **Transaction Management**: Distributed transaction support
- **Read/Write Splitting**: Separate read and write connections
- **Monitoring Dashboard**: Real-time pool statistics
- **Connection Retry Logic**: Automatic reconnection handling

## Troubleshooting

**Connection Failed**:
- Verify PostgreSQL is running
- Check connection string format
- Ensure network connectivity
- Verify database credentials

**Pool Exhausted**:
- Increase max_connections
- Check for connection leaks
- Monitor application load
- Consider connection timeouts

**Slow Queries**:
- Enable query logging
- Add database indexes
- Optimize query patterns
- Monitor connection pool metrics