# Configuration Examples

This folder contains three examples demonstrating different configuration patterns in Lokstra.

## Examples Overview

### [01-basic-yaml](./01-basic-yaml/)
**Basic YAML Configuration**
- Single YAML file configuration
- Simple routes and handlers
- Direct config loading

**Learn:** How to load and apply YAML configuration

---

### [02-environment-config](./02-environment-config/)
**Environment-Based Configuration**
- Base configuration + environment overrides
- Multi-file configuration merging
- Environment variables with defaults
- Runtime environment detection

**Learn:** How to manage different environments (dev, staging, prod)

---

### [03-cfg-references](./03-cfg-references/)
**CFG References (DRY Configuration)**
- Shared config values in `configs` section
- CFG reference syntax: `${@CFG:path.to.value}`
- Reuse values across configuration
- Dynamic path construction

**Learn:** How to avoid repetition and maintain consistency

---

## Quick Start

Each example is self-contained and can be run independently:

```bash
# Example 1 - Basic YAML
cd 01-basic-yaml
go run main.go

# Example 2 - Environment Config
cd 02-environment-config
go run main.go                    # Development
APP_ENV=prod go run main.go       # Production

# Example 3 - CFG References
cd 03-cfg-references
go run main.go
```

## Configuration Patterns Summary

| Pattern | Use Case | Example |
|---------|----------|---------|
| **Basic YAML** | Simple apps, single environment | Example 1 |
| **Environment-Based** | Multiple environments (dev/prod) | Example 2 |
| **CFG References** | Large configs, avoid repetition | Example 3 |
| **Combined** | Production apps (all patterns) | Mix all 3 |

## Best Practices

1. **Start Simple** - Begin with basic YAML (Example 1)
2. **Add Environments** - Use base + overrides (Example 2)
3. **Eliminate Duplication** - Use CFG references (Example 3)
4. **Code + Config** - Define logic in code, instances in YAML

## Next Steps

After completing these examples, check out:
- [05-app-and-server](../../05-app-and-server/) - Application lifecycle
- [06-putting-it-together](../../06-putting-it-together/) - Complete project
- [API Reference](../../../../api-reference/) - Full configuration options
