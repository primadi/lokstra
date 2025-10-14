# Quick Start Guide - Lokstra Auth System

## Setup (5 minutes)

### 1. Start PostgreSQL
```bash
# Using Docker (recommended for testing)
docker run -d \
  --name postgres-lokstra \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=lokstra_auth_demo \
  -p 5432:5432 \
  postgres:15-alpine

# Or use existing PostgreSQL installation
```

### 2. Start Redis
```bash
# Using Docker (recommended for testing)
docker run -d \
  --name redis-lokstra \
  -p 6379:6379 \
  redis:alpine

# Or use existing Redis installation
```

### 3. Initialize Database
```bash
# Copy and run the SQL script
psql -U postgres -d lokstra_auth_demo -f setup.sql

# Or manually create the users table (see setup.sql)
```

### 4. Run the Application
```bash
cd cmd/examples/17-auth-system
go run .
```

Server starts at: http://localhost:8080

## Quick Test (2 minutes)

### 1. Health Check
```bash
curl http://localhost:8080/api/health
```

### 2. Register a User
```bash
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "tenant_id": "tenant1",
    "username": "testuser",
    "email": "test@example.com",
    "full_name": "Test User",
    "password": "test123"
  }'
```

### 3. Login
```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "tenant_id": "tenant1",
    "username": "testuser",
    "password": "test123"
  }'
```

**Copy the `access_token` from the response!**

### 4. Access Protected Route
```bash
curl http://localhost:8080/api/user/profile \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

## Common Operations

### Create Admin User
```bash
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "tenant_id": "tenant1",
    "username": "admin",
    "email": "admin@example.com",
    "full_name": "Admin User",
    "password": "admin123",
    "role": "admin"
  }'
```

### List All Users (Admin Only)
```bash
curl http://localhost:8080/api/admin/users \
  -H "Authorization: Bearer ADMIN_ACCESS_TOKEN"
```

### Refresh Token
```bash
curl -X POST http://localhost:8080/api/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "YOUR_REFRESH_TOKEN"
  }'
```

## Architecture Overview

```
┌─────────────────────────────────────────────────────────┐
│                    HTTP Requests                        │
└────────────────────┬────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────┐
│              Middleware Chain                           │
│  ┌──────────┐  ┌──────────┐  ┌─────────────────┐      │
│  │   CORS   │→ │  Logger  │→ │    Recovery     │      │
│  └──────────┘  └──────────┘  └─────────────────┘      │
└────────────────────┬────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────┐
│                Route-Specific Middleware                │
│  ┌──────────────┐     ┌───────────────────┐           │
│  │   JwtAuth    │  →  │  AccessControl    │           │
│  │  (validate)  │     │  (check roles)    │           │
│  └──────────────┘     └───────────────────┘           │
└────────────────────┬────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────┐
│                    Handlers                             │
│  ┌─────────┐  ┌─────────┐  ┌────────┐  ┌─────────┐   │
│  │  Auth   │  │  User   │  │ Admin  │  │ Public  │   │
│  └─────────┘  └─────────┘  └────────┘  └─────────┘   │
└────────────────────┬────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────┐
│                    Services                             │
│  ┌──────────────┐  ┌────────────┐  ┌────────────┐     │
│  │ Auth Service │  │ User Repo  │  │  KvStore   │     │
│  └──────────────┘  └────────────┘  └────────────┘     │
└────────────────────┬────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────┐
│              Data Layer                                 │
│  ┌──────────────┐              ┌──────────────┐        │
│  │  PostgreSQL  │              │    Redis     │        │
│  │   (Users)    │              │  (Sessions)  │        │
│  └──────────────┘              └──────────────┘        │
└─────────────────────────────────────────────────────────┘
```

## Token Flow

```
1. User Login
   ┌──────┐                  ┌────────┐                ┌──────┐
   │Client│                  │ Server │                │ DB   │
   └───┬──┘                  └───┬────┘                └──┬───┘
       │ POST /auth/login       │                         │
       │ (username, password)   │                         │
       ├───────────────────────>│                         │
       │                        │ Verify credentials      │
       │                        ├────────────────────────>│
       │                        │<────────────────────────┤
       │                        │ Generate JWT tokens     │
       │                        │ Store refresh token     │
       │ {access_token,         │                         │
       │  refresh_token}        │                         │
       │<───────────────────────┤                         │
       │                        │                         │

2. Access Protected Resource
   ┌──────┐                  ┌────────┐
   │Client│                  │ Server │
   └───┬──┘                  └───┬────┘
       │ GET /user/profile      │
       │ Authorization: Bearer  │
       │    {access_token}      │
       ├───────────────────────>│
       │                        │ Validate token
       │                        │ Extract user info
       │                        │ Check permissions
       │ {user_data}            │
       │<───────────────────────┤
       │                        │

3. Refresh Token
   ┌──────┐                  ┌────────┐                ┌───────┐
   │Client│                  │ Server │                │ Redis │
   └───┬──┘                  └───┬────┘                └───┬───┘
       │ POST /auth/refresh     │                         │
       │ {refresh_token}        │                         │
       ├───────────────────────>│                         │
       │                        │ Validate refresh token  │
       │                        ├────────────────────────>│
       │                        │<────────────────────────┤
       │                        │ Generate new access token
       │                        │ Rotate refresh token    │
       │ {access_token,         │                         │
       │  refresh_token}        │                         │
       │<───────────────────────┤                         │
```

## File Structure

```
17-auth-system/
├── main.go              # Entry point
├── register.go          # Service & middleware registration
├── setup_routers.go     # Router configuration
├── handlers.go          # HTTP handlers
├── config.yaml          # Configuration
├── setup.sql            # Database schema
├── test-request.http    # API test requests
├── README.md            # Full documentation
└── QUICKSTART.md        # This file
```

## Configuration Notes

### Token Expiration
Edit `config.yaml`:
```yaml
auth_service:
  default:
    issuer:
      access_token_ttl: 900      # 15 minutes
      refresh_token_ttl: 604800  # 7 days
```

### Allowed Origins (CORS)
Edit `config.yaml`:
```yaml
middleware:
  cors:
    default:
      allowed_origins:
        - "http://localhost:3000"  # Add your frontend URL
```

### Database Connection
Edit `config.yaml`:
```yaml
dbpool_pg:
  default:
    host: localhost
    port: 5432
    user: postgres
    password: postgres
    dbname: lokstra_auth_demo
```

## Troubleshooting

### Server won't start
- Check if port 8080 is available
- Verify PostgreSQL is running
- Verify Redis is running
- Check config.yaml for correct settings

### "connection refused" errors
```bash
# Check PostgreSQL
docker ps | grep postgres
psql -U postgres -d lokstra_auth_demo -c "SELECT 1"

# Check Redis
docker ps | grep redis
redis-cli ping
```

### "invalid token" errors
- Token might be expired (15 min default)
- Use refresh token to get new access token
- Check Authorization header format: `Bearer <token>`

### "permission denied" errors
- Check user role in database
- Verify middleware is applied to route
- Admin routes require "admin" or "superadmin" role

## Next Steps

1. **Read the full README.md** for detailed documentation
2. **Test all endpoints** using test-request.http
3. **Customize for your needs**:
   - Add more roles
   - Implement email verification
   - Add OAuth2 support
   - Implement rate limiting
   - Add audit logging

## Support

For questions and issues:
- Check the full README.md
- Review other examples in `/cmd/examples/`
- Read Lokstra documentation

## Security Checklist

Before deploying to production:
- [ ] Change secret keys in config.yaml
- [ ] Use environment variables for secrets
- [ ] Enable HTTPS
- [ ] Implement rate limiting
- [ ] Add input validation
- [ ] Set up proper logging
- [ ] Configure CORS for production domains
- [ ] Use strong password policies
- [ ] Implement account lockout
- [ ] Add token rotation
- [ ] Set up monitoring and alerts
