# Kubernetes Secret Resolver

## Overview

The K8s Secret Resolver allows Lokstra to read configuration values from Kubernetes Secrets at runtime. This is perfect for:
- Running applications in Kubernetes clusters
- Storing sensitive data (passwords, API keys, certificates)
- Dynamic configuration without rebuilding containers
- GitOps workflows (secrets managed separately from config)

## Installation

Add Kubernetes client-go dependency:

```bash
go get k8s.io/client-go@latest
go get k8s.io/api@latest
go get k8s.io/apimachinery@latest
```

## Usage

### 1. Register the Resolver

```go
package main

import (
    "log"
    "github.com/primadi/lokstra/core/config"
    "github.com/primadi/lokstra/core/config/custom_resolver"
)

func main() {
    // Create K8s resolver (auto-detects in-cluster or kubeconfig)
    k8sResolver, err := custom_resolver.NewK8sSecretResolver()
    if err != nil {
        log.Fatal(err)
    }
    
    // Register resolver
    config.AddVariableResolver("K8S", k8sResolver)
    
    // Load your config (will use K8s secrets)
    cfg, err := config.LoadConfigFromFile("config/app.yaml")
    if err != nil {
        log.Fatal(err)
    }
    
    // Use config...
}
```

### 2. Create Kubernetes Secrets

```bash
# Database credentials
kubectl create secret generic database-credentials \
  --from-literal=host=postgres.default.svc.cluster.local \
  --from-literal=port=5432 \
  --from-literal=username=myuser \
  --from-literal=password=mypassword \
  --from-literal=database=mydb

# API keys
kubectl create secret generic api-secrets \
  --from-literal=api-key=sk_live_abc123 \
  --from-literal=jwt-secret=super-secret-jwt-key

# Redis auth
kubectl create secret generic redis-auth \
  --from-literal=url=redis://redis.default.svc.cluster.local:6379 \
  --from-literal=password=redis-password
```

### 3. Use in YAML Config

**config/production.yaml:**

```yaml
servers:
  - name: api-server
    # Read from K8s secret
    baseUrl: ${@K8S:app-config/base-url:http://localhost:8080}
    apps:
      - addr: :8080

services:
  # PostgreSQL with K8s secrets
  - name: postgres
    type: dbpool_pg
    config:
      # Default namespace (secret-name/key-name)
      host: ${@K8S:database-credentials/host}
      port: ${@K8S:database-credentials/port}
      user: ${@K8S:database-credentials/username}
      password: ${@K8S:database-credentials/password}
      database: ${@K8S:database-credentials/database}
  
  # Redis with K8s secrets
  - name: redis
    config:
      url: ${@K8S:redis-auth/url}
      password: ${@K8S:redis-auth/password}
  
  # API with secrets from specific namespace
  - name: api
    config:
      # Explicit namespace (namespace/secret-name/key-name)
      apiKey: ${@K8S:production/api-secrets/api-key}
      jwtSecret: ${@K8S:production/api-secrets/jwt-secret}
```

## Key Format

The K8s resolver supports two key formats:

### Format 1: Default Namespace
```
secret-name/key-name
```

Example: `database-credentials/password`
- Uses default namespace (from `POD_NAMESPACE` env or "default")

### Format 2: Explicit Namespace
```
namespace/secret-name/key-name
```

Example: `production/api-secrets/api-key`
- Reads from specific namespace

## Authentication

The resolver automatically detects the environment:

### In-Cluster (Recommended for Production)
When running inside Kubernetes:
- Uses ServiceAccount token automatically
- No configuration needed
- Most secure option

```yaml
# deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-app
spec:
  template:
    spec:
      serviceAccountName: my-app  # Uses this ServiceAccount
      containers:
      - name: app
        image: my-app:latest
        env:
        - name: POD_NAMESPACE
          valueFrom:
            fieldRef:
              fieldPath: metadata.namespace
```

### Kubeconfig (Development)
When running locally:
- Uses `~/.kube/config` automatically
- Or `KUBECONFIG` environment variable
- Great for local development

```bash
# Set custom kubeconfig
export KUBECONFIG=/path/to/kubeconfig

# Run your app
go run main.go
```

## RBAC Permissions

Your ServiceAccount needs permission to read secrets:

**rbac.yaml:**

```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: my-app
  namespace: default
---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: secret-reader
  namespace: default
rules:
- apiGroups: [""]
  resources: ["secrets"]
  verbs: ["get", "list"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: my-app-secret-reader
  namespace: default
subjects:
- kind: ServiceAccount
  name: my-app
  namespace: default
roleRef:
  kind: Role
  name: secret-reader
  apiGroup: rbac.authorization.k8s.io
```

Apply RBAC:
```bash
kubectl apply -f rbac.yaml
```

## Advanced Features

### Custom Default Namespace

```go
k8sResolver, err := custom_resolver.NewK8sSecretResolverWithNamespace("production")
```

### Retrieve Entire Secret

```go
secretData, err := k8sResolver.ResolveEntireSecret("default", "database-credentials")
// Returns: map[string]string{
//   "host": "postgres.default.svc.cluster.local",
//   "port": "5432",
//   "username": "myuser",
//   "password": "mypassword",
//   "database": "mydb",
// }
```

### List All Secrets

```go
secrets, err := k8sResolver.ListSecrets("default")
// Returns: ["database-credentials", "api-secrets", "redis-auth"]
```

### Watch Secret Changes (Real-time Updates)

```go
ch, err := k8sResolver.WatchSecret("default", "api-secrets")
if err != nil {
    log.Fatal(err)
}

go func() {
    for secret := range ch {
        log.Printf("Secret updated: %s", secret.Name)
        // Clear cache to pick up new values
        k8sResolver.ClearCache()
    }
}()
```

### Clear Cache

```go
// Clear cache to force re-read from K8s
k8sResolver.ClearCache()
```

## Best Practices

### 1. Use Fallback Defaults

Always provide defaults for local development:

```yaml
database:
  # Prod: reads from K8s secret
  # Dev: uses localhost default
  host: ${@K8S:database-credentials/host:localhost}
  port: ${@K8S:database-credentials/port:5432}
```

### 2. Separate Secrets by Environment

```yaml
# Development
password: ${@K8S:dev/database-credentials/password:dev-password}

# Staging
password: ${@K8S:staging/database-credentials/password}

# Production
password: ${@K8S:production/database-credentials/password}
```

### 3. Use ServiceAccount per App

```yaml
# One ServiceAccount per application
apiVersion: v1
kind: ServiceAccount
metadata:
  name: api-app
  namespace: production
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: worker-app
  namespace: production
```

### 4. Minimal RBAC Permissions

Only grant access to secrets your app needs:

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: api-app-secrets
rules:
- apiGroups: [""]
  resources: ["secrets"]
  resourceNames: ["database-credentials", "api-secrets"]  # Specific secrets only
  verbs: ["get"]
```

### 5. Use POD_NAMESPACE Environment Variable

```yaml
env:
- name: POD_NAMESPACE
  valueFrom:
    fieldRef:
      fieldPath: metadata.namespace
```

This ensures resolver uses correct namespace automatically.

## Examples

### Example 1: Database Connection

**Kubernetes Secret:**
```bash
kubectl create secret generic db-creds \
  --from-literal=dsn="postgresql://user:pass@postgres:5432/mydb?sslmode=require"
```

**Config:**
```yaml
services:
  - name: postgres
    type: dbpool_pg
    config:
      dsn: ${@K8S:db-creds/dsn:postgresql://localhost:5432/dev}
```

### Example 2: API Keys

**Kubernetes Secret:**
```bash
kubectl create secret generic api-keys \
  --from-literal=stripe-key=sk_live_abc123 \
  --from-literal=sendgrid-key=SG.xyz789
```

**Config:**
```yaml
services:
  - name: api
    config:
      stripeKey: ${@K8S:api-keys/stripe-key}
      sendgridKey: ${@K8S:api-keys/sendgrid-key}
```

### Example 3: TLS Certificates

**Kubernetes Secret:**
```bash
kubectl create secret tls api-tls \
  --cert=./tls.crt \
  --key=./tls.key
```

**Config:**
```yaml
servers:
  - name: api
    tls:
      certFile: ${@K8S:api-tls/tls.crt}
      keyFile: ${@K8S:api-tls/tls.key}
```

### Example 4: Multi-Namespace Setup

**Kubernetes Secrets:**
```bash
# Shared database credentials
kubectl create secret generic db-creds \
  --namespace=shared \
  --from-literal=host=postgres.shared.svc.cluster.local \
  --from-literal=password=shared-password

# App-specific API key
kubectl create secret generic api-key \
  --namespace=app-production \
  --from-literal=key=app-specific-key
```

**Config:**
```yaml
services:
  - name: postgres
    type: dbpool_pg
    config:
      host: ${@K8S:shared/db-creds/host}
      password: ${@K8S:shared/db-creds/password}
  
  - name: api
    config:
      apiKey: ${@K8S:app-production/api-key/key}
```

## Troubleshooting

### Error: "failed to load kubernetes config"

**Cause**: Not running in K8s cluster and no kubeconfig found

**Solution**:
```bash
# Check kubeconfig exists
ls ~/.kube/config

# Or set KUBECONFIG
export KUBECONFIG=/path/to/config

# Test kubectl works
kubectl get pods
```

### Error: "secrets is forbidden"

**Cause**: ServiceAccount lacks RBAC permissions

**Solution**:
```bash
# Check current permissions
kubectl auth can-i get secrets --as=system:serviceaccount:default:my-app

# Apply RBAC (see RBAC Permissions section above)
kubectl apply -f rbac.yaml
```

### Error: "secret not found"

**Cause**: Secret doesn't exist or wrong namespace

**Solution**:
```bash
# List secrets
kubectl get secrets -n default

# Describe secret
kubectl describe secret database-credentials -n default

# Check key exists
kubectl get secret database-credentials -o jsonpath='{.data.password}' | base64 -d
```

### Cache Issues

If secrets are updated but app still uses old values:

```go
// Clear cache manually
k8sResolver.ClearCache()

// Or restart pod to clear cache
kubectl rollout restart deployment/my-app
```

## Comparison with Other Secret Management

| Feature | K8s Secrets | AWS Secrets | Vault |
|---------|-------------|-------------|-------|
| **Native K8s** | ✅ Yes | ❌ No | ❌ No |
| **Setup Complexity** | ✅ Simple | ⚠️ Medium | ⚠️ Complex |
| **Cost** | ✅ Free | 💰 Paid | ✅ Free (OSS) |
| **Rotation** | ⚠️ Manual | ✅ Auto | ✅ Auto |
| **Audit Logs** | ⚠️ Basic | ✅ CloudTrail | ✅ Audit Log |
| **Encryption** | ✅ etcd encrypted | ✅ KMS | ✅ Transit |
| **Best For** | K8s-native apps | AWS apps | Multi-cloud |

**Recommendation**: 
- Use K8s Secrets for **Kubernetes-native applications**
- Use AWS Secrets for **AWS-heavy workloads**
- Use Vault for **multi-cloud** or **complex rotation** needs

## See Also

- [Environment Variable Syntax](../../../docs/environment-variable-syntax.md)
- [AWS Secrets Resolver](./aws_secret.go)
- [Custom Resolver Examples](../../../docs/examples/custom-resolvers.go)
- [Kubernetes Secrets Documentation](https://kubernetes.io/docs/concepts/configuration/secret/)
