# Kubernetes Deployment Example

This example shows how to deploy a Lokstra application to Kubernetes with secret management.

## Directory Structure

```
k8s/
├── namespace.yaml           # Namespace definition
├── secrets.yaml            # Application secrets
├── rbac.yaml              # ServiceAccount and RBAC
├── configmap.yaml         # Application config (non-sensitive)
├── deployment.yaml        # Application deployment
└── service.yaml           # Service exposure
```

## Step-by-Step Deployment

### 1. Create Namespace

**namespace.yaml:**
```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: lokstra-app
  labels:
    app: lokstra
    environment: production
```

Apply:
```bash
kubectl apply -f namespace.yaml
```

---

### 2. Create Secrets

**secrets.yaml:**
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: database-credentials
  namespace: lokstra-app
type: Opaque
stringData:
  host: postgres.lokstra-app.svc.cluster.local
  port: "5432"
  username: lokstra_user
  password: supersecretpassword123
  database: lokstra_db
---
apiVersion: v1
kind: Secret
metadata:
  name: api-secrets
  namespace: lokstra-app
type: Opaque
stringData:
  api-key: sk_live_1234567890abcdef
  jwt-secret: jwt-super-secret-key-change-in-production
  stripe-key: sk_live_stripe_abc123xyz
---
apiVersion: v1
kind: Secret
metadata:
  name: redis-auth
  namespace: lokstra-app
type: Opaque
stringData:
  url: redis://redis.lokstra-app.svc.cluster.local:6379
  password: redis-password-123
```

Apply:
```bash
kubectl apply -f secrets.yaml
```

**⚠️ Security Note**: In production, use `kubectl create secret` instead of YAML files:

```bash
# Database credentials
kubectl create secret generic database-credentials \
  --namespace=lokstra-app \
  --from-literal=host=postgres.lokstra-app.svc.cluster.local \
  --from-literal=port=5432 \
  --from-literal=username=lokstra_user \
  --from-literal=password=$(openssl rand -base64 32) \
  --from-literal=database=lokstra_db

# API secrets
kubectl create secret generic api-secrets \
  --namespace=lokstra-app \
  --from-literal=api-key=sk_live_1234567890abcdef \
  --from-literal=jwt-secret=$(openssl rand -base64 32) \
  --from-literal=stripe-key=sk_live_stripe_abc123xyz

# Redis auth
kubectl create secret generic redis-auth \
  --namespace=lokstra-app \
  --from-literal=url=redis://redis.lokstra-app.svc.cluster.local:6379 \
  --from-literal=password=$(openssl rand -base64 32)
```

---

### 3. Create RBAC

**rbac.yaml:**
```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: lokstra-app
  namespace: lokstra-app
  labels:
    app: lokstra
---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: secret-reader
  namespace: lokstra-app
rules:
- apiGroups: [""]
  resources: ["secrets"]
  resourceNames:
    - database-credentials
    - api-secrets
    - redis-auth
  verbs: ["get"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: lokstra-app-secret-reader
  namespace: lokstra-app
subjects:
- kind: ServiceAccount
  name: lokstra-app
  namespace: lokstra-app
roleRef:
  kind: Role
  name: secret-reader
  apiGroup: rbac.authorization.k8s.io
```

Apply:
```bash
kubectl apply -f rbac.yaml
```

---

### 4. Create ConfigMap (Optional)

For non-sensitive configuration:

**configmap.yaml:**
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: lokstra-config
  namespace: lokstra-app
data:
  # Lokstra YAML config
  app.yaml: |
    servers:
      - name: api-server
        # Read from K8s secret with fallback
        baseUrl: ${@K8S:app-config/base-url:http://localhost:8080}
        apps:
          - addr: :8080
            name: api
    
    services:
      # PostgreSQL - all credentials from secrets
      - name: postgres
        type: dbpool_pg
        enable: true
        config:
          host: ${@K8S:database-credentials/host}
          port: ${@K8S:database-credentials/port}
          user: ${@K8S:database-credentials/username}
          password: ${@K8S:database-credentials/password}
          database: ${@K8S:database-credentials/database}
          maxConns: 25
          minConns: 5
      
      # Redis - credentials from secrets
      - name: redis
        enable: true
        config:
          url: ${@K8S:redis-auth/url}
          password: ${@K8S:redis-auth/password}
      
      # API configuration - keys from secrets
      - name: api
        config:
          apiKey: ${@K8S:api-secrets/api-key}
          jwtSecret: ${@K8S:api-secrets/jwt-secret}
          stripeKey: ${@K8S:api-secrets/stripe-key}
          
    routers:
      - name: main-router
        basePath: /api/v1
```

Apply:
```bash
kubectl apply -f configmap.yaml
```

---

### 5. Create Deployment

**deployment.yaml:**
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: lokstra-app
  namespace: lokstra-app
  labels:
    app: lokstra
    version: v1.0.0
spec:
  replicas: 3
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 0
  selector:
    matchLabels:
      app: lokstra
  template:
    metadata:
      labels:
        app: lokstra
        version: v1.0.0
      annotations:
        # Force restart on config change
        checksum/config: {{ include (print $.Template.BasePath "/configmap.yaml") . | sha256sum }}
    spec:
      serviceAccountName: lokstra-app
      
      # Init container to check dependencies
      initContainers:
      - name: wait-for-postgres
        image: busybox:1.36
        command: 
        - sh
        - -c
        - |
          until nc -z postgres.lokstra-app.svc.cluster.local 5432; do
            echo "Waiting for PostgreSQL..."
            sleep 2
          done
          echo "PostgreSQL is ready!"
      
      containers:
      - name: app
        image: your-registry/lokstra-app:v1.0.0
        imagePullPolicy: Always
        
        ports:
        - name: http
          containerPort: 8080
          protocol: TCP
        
        env:
        # Tell resolver which namespace to use
        - name: POD_NAMESPACE
          valueFrom:
            fieldRef:
              fieldPath: metadata.namespace
        
        # App environment
        - name: ENVIRONMENT
          value: production
        
        # Config file location
        - name: CONFIG_FILE
          value: /etc/lokstra/app.yaml
        
        # Optional: Log level
        - name: LOG_LEVEL
          value: info
        
        # Volume mounts
        volumeMounts:
        - name: config
          mountPath: /etc/lokstra
          readOnly: true
        
        # Health checks
        livenessProbe:
          httpGet:
            path: /health/live
            port: http
          initialDelaySeconds: 30
          periodSeconds: 10
          timeoutSeconds: 5
          failureThreshold: 3
        
        readinessProbe:
          httpGet:
            path: /health/ready
            port: http
          initialDelaySeconds: 10
          periodSeconds: 5
          timeoutSeconds: 3
          failureThreshold: 2
        
        # Resource limits
        resources:
          requests:
            cpu: 100m
            memory: 128Mi
          limits:
            cpu: 500m
            memory: 512Mi
        
        # Security context
        securityContext:
          allowPrivilegeEscalation: false
          runAsNonRoot: true
          runAsUser: 1000
          capabilities:
            drop:
            - ALL
      
      # Volumes
      volumes:
      - name: config
        configMap:
          name: lokstra-config
      
      # Pod security
      securityContext:
        fsGroup: 1000
        runAsNonRoot: true
        seccompProfile:
          type: RuntimeDefault
```

Apply:
```bash
kubectl apply -f deployment.yaml
```

---

### 6. Create Service

**service.yaml:**
```yaml
apiVersion: v1
kind: Service
metadata:
  name: lokstra-app
  namespace: lokstra-app
  labels:
    app: lokstra
spec:
  type: ClusterIP
  selector:
    app: lokstra
  ports:
  - name: http
    port: 80
    targetPort: http
    protocol: TCP
  sessionAffinity: None
---
# Optional: Ingress
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: lokstra-app
  namespace: lokstra-app
  annotations:
    cert-manager.io/cluster-issuer: letsencrypt-prod
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
spec:
  ingressClassName: nginx
  tls:
  - hosts:
    - api.example.com
    secretName: lokstra-app-tls
  rules:
  - host: api.example.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: lokstra-app
            port:
              name: http
```

Apply:
```bash
kubectl apply -f service.yaml
```

---

## Complete Deployment Script

**deploy.sh:**
```bash
#!/bin/bash

set -e

NAMESPACE="lokstra-app"
IMAGE_TAG=${1:-latest}

echo "🚀 Deploying Lokstra App to Kubernetes"
echo "Namespace: $NAMESPACE"
echo "Image Tag: $IMAGE_TAG"
echo ""

# 1. Create namespace
echo "📁 Creating namespace..."
kubectl apply -f namespace.yaml

# 2. Create secrets (if not exists)
if ! kubectl get secret database-credentials -n $NAMESPACE &>/dev/null; then
    echo "🔐 Creating secrets..."
    kubectl create secret generic database-credentials \
      --namespace=$NAMESPACE \
      --from-literal=host=postgres.lokstra-app.svc.cluster.local \
      --from-literal=port=5432 \
      --from-literal=username=lokstra_user \
      --from-literal=password=$(openssl rand -base64 32) \
      --from-literal=database=lokstra_db
    
    kubectl create secret generic api-secrets \
      --namespace=$NAMESPACE \
      --from-literal=api-key=sk_live_$(openssl rand -hex 16) \
      --from-literal=jwt-secret=$(openssl rand -base64 32) \
      --from-literal=stripe-key=sk_live_stripe_$(openssl rand -hex 16)
    
    kubectl create secret generic redis-auth \
      --namespace=$NAMESPACE \
      --from-literal=url=redis://redis.lokstra-app.svc.cluster.local:6379 \
      --from-literal=password=$(openssl rand -base64 32)
else
    echo "✓ Secrets already exist"
fi

# 3. Apply RBAC
echo "👤 Setting up RBAC..."
kubectl apply -f rbac.yaml

# 4. Apply ConfigMap
echo "⚙️  Applying configuration..."
kubectl apply -f configmap.yaml

# 5. Update deployment image
echo "🐳 Updating deployment..."
sed "s|image: .*lokstra-app:.*|image: your-registry/lokstra-app:$IMAGE_TAG|g" deployment.yaml | kubectl apply -f -

# 6. Apply service
echo "🌐 Setting up service..."
kubectl apply -f service.yaml

# 7. Wait for rollout
echo "⏳ Waiting for rollout..."
kubectl rollout status deployment/lokstra-app -n $NAMESPACE --timeout=5m

# 8. Show status
echo ""
echo "✅ Deployment complete!"
echo ""
kubectl get pods -n $NAMESPACE -l app=lokstra
echo ""
kubectl get svc -n $NAMESPACE -l app=lokstra
```

Make executable and run:
```bash
chmod +x deploy.sh
./deploy.sh v1.0.0
```

---

## Verify Deployment

```bash
# Check pods
kubectl get pods -n lokstra-app

# Check logs
kubectl logs -n lokstra-app -l app=lokstra --tail=100 -f

# Check secrets are accessible
kubectl exec -n lokstra-app deployment/lokstra-app -- \
  sh -c 'echo "Testing secret access..."'

# Port forward for local testing
kubectl port-forward -n lokstra-app svc/lokstra-app 8080:80

# Test endpoint
curl http://localhost:8080/health
```

---

## Update Secrets

When you need to update secrets:

```bash
# Update database password
kubectl create secret generic database-credentials \
  --namespace=lokstra-app \
  --from-literal=password=new-password-here \
  --dry-run=client -o yaml | kubectl apply -f -

# Restart pods to pick up new secrets
kubectl rollout restart deployment/lokstra-app -n lokstra-app
```

---

## Troubleshooting

### Pods are CrashLooping

```bash
# Check pod logs
kubectl logs -n lokstra-app -l app=lokstra --tail=200

# Describe pod
kubectl describe pod -n lokstra-app -l app=lokstra

# Check secret exists
kubectl get secrets -n lokstra-app
```

### Cannot Access Secrets

```bash
# Check RBAC permissions
kubectl auth can-i get secrets \
  --namespace=lokstra-app \
  --as=system:serviceaccount:lokstra-app:lokstra-app

# Check ServiceAccount is correct
kubectl get pod -n lokstra-app -l app=lokstra -o yaml | grep serviceAccountName
```

### Secret Not Found

```bash
# List all secrets
kubectl get secrets -n lokstra-app

# Check secret data
kubectl get secret database-credentials -n lokstra-app -o yaml

# Decode secret value
kubectl get secret database-credentials -n lokstra-app \
  -o jsonpath='{.data.password}' | base64 -d
```

---

## Best Practices

1. **Never commit secrets to Git**
   - Use `.gitignore` for `secrets.yaml`
   - Create secrets via `kubectl create secret` command

2. **Use sealed-secrets or external-secrets**
   - For GitOps workflows
   - [Sealed Secrets](https://github.com/bitnami-labs/sealed-secrets)
   - [External Secrets Operator](https://external-secrets.io/)

3. **Rotate secrets regularly**
   ```bash
   # Automate secret rotation
   kubectl create secret generic database-credentials \
     --from-literal=password=$(openssl rand -base64 32) \
     --dry-run=client -o yaml | kubectl apply -f -
   
   kubectl rollout restart deployment/lokstra-app -n lokstra-app
   ```

4. **Use separate namespaces per environment**
   ```
   lokstra-dev
   lokstra-staging
   lokstra-production
   ```

5. **Enable encryption at rest**
   - Configure etcd encryption
   - [Kubernetes Encryption at Rest](https://kubernetes.io/docs/tasks/administer-cluster/encrypt-data/)

---

## See Also

- [Kubernetes Secret Resolver README](./README.md)
- [Lokstra Configuration Guide](../../../docs/yaml-configuration-system.md)
- [Environment Variable Syntax](../../../docs/environment-variable-syntax.md)
