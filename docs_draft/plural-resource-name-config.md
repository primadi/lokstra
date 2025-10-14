# Plural Resource Name Configuration

## Overview

The `plural-resource-name` field allows you to override the automatic pluralization of resource names in REST conventions. This is useful for irregular plurals or language-specific variations.

## Configuration

### YAML Config

```yaml
services:
  - name: person-service
    type: person_service
    auto-router:
      convention: rest
      resource-name: person
      plural-resource-name: people  # Override default "persons"
```

### Without Override (Auto-Pluralization)

```yaml
services:
  - name: user-service
    type: user_service
    auto-router:
      convention: rest
      resource-name: user
      # plural-resource-name not specified
      # Will auto-pluralize to "users"
```

## Generated Routes

### With `plural-resource-name: people`

```
GET    /people        → ListPersons()
GET    /people/{id}   → GetPerson()
POST   /people        → CreatePerson()
PUT    /people/{id}   → UpdatePerson()
DELETE /people/{id}   → DeletePerson()
PATCH  /people/{id}   → PatchPerson()
```

### Without override (auto: "users")

```
GET    /users         → ListUsers()
GET    /users/{id}    → GetUser()
POST   /users         → CreateUser()
PUT    /users/{id}    → UpdateUser()
DELETE /users/{id}    → DeleteUser()
PATCH  /users/{id}    → PatchUser()
```

## Common Use Cases

### Irregular Plurals

```yaml
# person → people (not "persons")
- name: person-service
  auto-router:
    resource-name: person
    plural-resource-name: people

# child → children (not "childs")
- name: child-service
  auto-router:
    resource-name: child
    plural-resource-name: children

# tooth → teeth (not "tooths")
- name: tooth-service
  auto-router:
    resource-name: tooth
    plural-resource-name: teeth
```

### Uncountable Nouns

```yaml
# data (already plural)
- name: data-service
  auto-router:
    resource-name: data
    plural-resource-name: data

# information (uncountable)
- name: info-service
  auto-router:
    resource-name: information
    plural-resource-name: information
```

### Non-English Resources

```yaml
# Indonesia: "produk" (both singular and plural)
- name: product-service
  auto-router:
    resource-name: produk
    plural-resource-name: produk
```

## Priority

The system follows this priority order:

1. **`AutoRouter.PluralResourceName`** (highest) - Explicit override from config
2. **Auto-pluralization** - Computed from `ResourceName` using simple rules:
   - Ends with 's', 'x', 'ch' → add 'es' (box → boxes)
   - Ends with 'y' → replace with 'ies' (city → cities)
   - Default → add 's' (user → users)

## Programmatic API

You can also set this when creating routers programmatically:

```go
options := router.DefaultServiceRouterOptions().
    WithConvention("rest").
    WithResourceName("person").
    WithPluralResourceName("people")

personRouter := router.NewFromService(personService, options)
```

## Helper Methods

### Service Config Methods

```go
// Get plural resource name (returns empty if not set)
pluralName := service.GetPluralResourceName()

// Get resource name (with fallbacks)
resourceName := service.GetResourceName()
```

### In Convention Implementation

```go
// In RESTConvention.GenerateRoutes()
resourceName := options.ResourceName    // "person"
pluralName := options.PluralResourceName // "people" or empty

if pluralName == "" {
    pluralName = pluralize(resourceName) // Auto-compute
}
```

## Related

- [Resource Name Configuration](./resource-name-config.md)
- [Auto-Router Configuration](./auto-router-config.md)
- [REST Convention](./convention-rest.md)
