# Authentication Module

This folder contains the auth module implementation for multi-tenant authentication.

## Module Structure

```
modules/auth/
├── handler.go          # HTTP handlers with @Handler annotations
├── service.go          # Business logic with @Service annotations  
├── repository.go       # Database operations with @Service annotations
├── dto.go              # Request/Response DTOs with validation
├── entity.go           # Domain entities
├── errors.go           # Module-specific error codes
└── README.md           # This file
```

## Quick Start

After implementing this module, you can use these endpoints:

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/auth/register` | Register new user (admin only) |
| POST | `/api/v1/auth/login` | User login |
| POST | `/api/v1/auth/logout` | User logout |
| POST | `/api/v1/auth/refresh` | Refresh access token |
| POST | `/api/v1/auth/validate` | Validate token (internal) |
| GET | `/api/v1/auth/me` | Get current user |
| POST | `/api/v1/auth/password/forgot` | Request password reset |
| POST | `/api/v1/auth/password/reset` | Reset password |
| POST | `/api/v1/auth/password/change` | Change password |
| POST | `/api/v1/auth/switch-tenant` | Switch tenant (super admin) |

## Requirements Document

See [docs/modules/auth/REQUIREMENTS-auth-v1.0.md](../../docs/modules/auth/REQUIREMENTS-auth-v1.0.md) for:
- Functional requirements
- Domain model
- Use cases
- Validation rules
- Security requirements
- Multi-tenant considerations

## Implementation Guide

Use Lokstra AI skills to implement:
1. `lokstra-api-specification` - Generate OpenAPI spec
2. `lokstra-schema-design` - Design database schema  
3. `implementation-lokstra-create-handler` - Create handlers
4. `implementation-lokstra-create-service` - Create services
5. `advanced-lokstra-tests` - Create tests

## Multi-Tenant Notes

- All queries MUST include `tenant_id` filter
- JWT tokens contain `tenant_id` claim
- Email unique per tenant (same email OK across tenants)
- Username globally unique
- Super admin can switch tenants
