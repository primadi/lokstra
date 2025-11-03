# Secrets Management

Secure handling of sensitive configuration data.

## Running

```bash
DB_PASSWORD=secret123 \
API_KEY=apikey456 \
JWT_SECRET=jwtsecret789 \
go run main.go
```

Server starts on `http://localhost:3050`

## Best Practices

### ✅ DO

- Store secrets in environment variables
- Use secret management services (Vault, AWS Secrets Manager)
- Rotate secrets regularly
- Use different secrets per environment
- Validate secrets at startup

### ❌ DON'T

- Commit secrets to version control
- Log secret values
- Expose secrets in API responses
- Hardcode secrets in configuration files
- Share production secrets

## Environment Variables

- `DB_PASSWORD` - Database password
- `API_KEY` - External API key
- `JWT_SECRET` - JWT signing secret

## Files

- `.env.example` - Template for environment variables
- `.env.local` - Your local secrets (gitignored)
