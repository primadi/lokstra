# YAML Configuration Examples

This folder contains practical examples demonstrating various aspects of YAML configuration in Lokstra.

## 📂 Examples

### [01 - Basic Configuration](01-basic-config/)
**What you'll learn:**
- Single-file YAML configuration
- Service definitions with dependencies
- Router auto-generation
- Middleware configuration
- Basic deployment setup

**Best for:** Getting started with YAML config

---

### [02 - Multi-File Configuration](02-multi-file/)
**What you'll learn:**
- Splitting config across multiple files
- Base + environment-specific configs
- Config merging strategy
- Development vs production setups

**Best for:** Managing multiple environments

---

### [06 - Handler Configurations](06-handlers/)
**What you'll learn:**
- Reverse proxy configuration
- Path stripping and rewriting
- SPA mounting
- Static file serving
- API gateway patterns

**Best for:** Infrastructure-level routing

---

### [07 - Named Database Pools](07-db-pools/)
**What you'll learn:**
- Multiple database pool configuration
- DSN vs component-based config
- Pool sizing and optimization
- SSL configuration
- Service dependencies on specific pools

**Best for:** Multi-database applications

---

## 🚀 Quick Start

Each example contains:
- `config.yaml` - YAML configuration file(s)
- `main.go` - Application entry point
- `README.md` - Detailed explanation

To run any example:
```bash
cd <example-folder>
go run main.go
```

## 📚 Learning Path

Recommended order:

1. **01-basic-config** → Understand fundamentals
2. **02-multi-file** → Learn config organization
3. **06-handlers** → Master infrastructure handlers
4. **07-db-pools** → Configure databases

## 💡 Key Concepts

### Configuration Hierarchy
```
configs                      # Global config values
dbpool-definitions              # Database pool definitions
middleware-definitions      # Middleware instances
service-definitions         # Service instances
router-definitions         # Router configurations
external-service-definitions # Remote services
deployments                # Deployment topologies
  └─ servers               # Server instances
      └─ apps              # Application instances
```

### Auto-Generation Flow
```
1. @Handler annotation → Service metadata
2. published-services → Auto-generate router
3. service.router → Router customization
4. router-definitions → Override defaults
```

### Handler Mount Order
```
1. Reverse Proxies (prepended first)
2. Business Routers (routers + published-services)
3. SPA Mounts
4. Static Mounts
```

## 🎯 Common Patterns

### Minimal Config (Convention over Configuration)
```yaml
service-definitions:
  user-service:
    type: user-service-factory
    depends-on: [user-repository]
    router: {}  # Uses conventions from @Handler

deployments:
  dev:
    servers:
      api:
        addr: ":8080"
        published-services: [user-service]
```

### Explicit Config (Full Control)
```yaml
service-definitions:
  user-service:
    type: user-service-factory
    depends-on: [user-repository]
    router:
      convention: rest
      resource: user
      path-prefix: /api/v1
      middlewares: [cors, auth]
      custom:
        - name: GetByEmail
          path: /by-email/{email}
```

## 🔗 Related Documentation

- [YAML Configuration Guide](../) - Full documentation
- [Service Deep Dive](../../02-service/) - Service architecture
- [Middleware Deep Dive](../../03-middleware/) - Custom middleware
- [App & Server Deep Dive](../../05-app-and-server/) - Production deployment

## 💬 Need Help?

- Check the main [Configuration Guide](../)
- Review example READMEs for detailed explanations
- Look at the [Quick Reference](../../../../QUICK-REFERENCE.md)

---

**Ready to dive in?** Start with [01-basic-config](01-basic-config/) →
