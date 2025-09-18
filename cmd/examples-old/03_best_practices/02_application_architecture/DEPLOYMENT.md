# Deployment Guide

This guide shows how to deploy the User Management application using various methods.

## Prerequisites

- Go 1.21+
- PostgreSQL database
- Git (for cloning)

## Method 1: Direct Deployment

### 1. Clone and Setup
```bash
git clone <your-repo>
cd cmd/examples/03_best_practices/02_application_architecture

# Install dependencies
go mod tidy
```

### 2. Database Setup
```bash
# Create database
createdb lokstra_example

# Run migrations
psql lokstra_example < migrations/001_create_users_table.sql
```

### 3. Configuration
Update `lokstra.yaml` with your database settings:
```yaml
services:
  - name: "db_pool"
    type: "lokstra.dbpool.pg"
    config:
      host: "your-db-host"
      port: 5432
      database: "your-database"
      username: "your-username"
      password: "your-password"
```

### 4. Run Application
```bash
go run main.go
```

## Method 2: Using Docker

### 1. Create Dockerfile
```dockerfile
# Multi-stage build
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o main main.go

# Final stage
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/main .
COPY --from=builder /app/lokstra.yaml .
COPY --from=builder /app/migrations/ ./migrations/

CMD ["./main"]
```

### 2. Create docker-compose.yml
```yaml
version: '3.8'
services:
  app:
    build: .
    ports:
      - "8080:8080"
    environment:
      - DB_HOST=postgres
      - DB_USER=postgres
      - DB_PASSWORD=password
      - DB_NAME=lokstra_example
    depends_on:
      - postgres

  postgres:
    image: postgres:15
    environment:
      - POSTGRES_USER=postgres
      - POSTGRES_PASSWORD=password
      - POSTGRES_DB=lokstra_example
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./migrations:/docker-entrypoint-initdb.d/
    ports:
      - "5432:5432"

volumes:
  postgres_data:
```

### 3. Deploy with Docker
```bash
docker-compose up -d
```

## Method 3: Kubernetes Deployment

### 1. Create Kubernetes Manifests

**namespace.yaml:**
```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: lokstra-example
```

**configmap.yaml:**
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: app-config
  namespace: lokstra-example
data:
  lokstra.yaml: |
    server:
      name: "application_architecture_example"
      global_setting:
        log_level: "info"
    
    services:
      - name: "logger"
        type: "lokstra.logger"
        config:
          level: "info"
          format: "json"
          output: "stdout"
    
      - name: "db_pool"
        type: "lokstra.dbpool.pg"
        config:
          host: "postgres-service"
          port: 5432
          database: "lokstra_example"
          username: "postgres"
          password: "password"
          max_connections: 25
          min_connections: 5
    
    apps:
      - name: "main_app"
        address: ":8080"
        groups:
          - prefix: "/api"
            routes:
              - method: "GET"
                path: "/health"
                handler: "health.check"
```

**deployment.yaml:**
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: lokstra-app
  namespace: lokstra-example
spec:
  replicas: 3
  selector:
    matchLabels:
      app: lokstra-app
  template:
    metadata:
      labels:
        app: lokstra-app
    spec:
      containers:
      - name: app
        image: your-registry/lokstra-example:latest
        ports:
        - containerPort: 8080
        volumeMounts:
        - name: config
          mountPath: /app/lokstra.yaml
          subPath: lokstra.yaml
        env:
        - name: PORT
          value: "8080"
      volumes:
      - name: config
        configMap:
          name: app-config
```

**service.yaml:**
```yaml
apiVersion: v1
kind: Service
metadata:
  name: lokstra-app-service
  namespace: lokstra-example
spec:
  selector:
    app: lokstra-app
  ports:
  - port: 80
    targetPort: 8080
  type: LoadBalancer
```

**postgres.yaml:**
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: postgres
  namespace: lokstra-example
spec:
  replicas: 1
  selector:
    matchLabels:
      app: postgres
  template:
    metadata:
      labels:
        app: postgres
    spec:
      containers:
      - name: postgres
        image: postgres:15
        env:
        - name: POSTGRES_DB
          value: lokstra_example
        - name: POSTGRES_USER
          value: postgres
        - name: POSTGRES_PASSWORD
          value: password
        ports:
        - containerPort: 5432
        volumeMounts:
        - name: postgres-storage
          mountPath: /var/lib/postgresql/data
      volumes:
      - name: postgres-storage
        persistentVolumeClaim:
          claimName: postgres-pvc

---
apiVersion: v1
kind: Service
metadata:
  name: postgres-service
  namespace: lokstra-example
spec:
  selector:
    app: postgres
  ports:
  - port: 5432
    targetPort: 5432
  type: ClusterIP

---
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: postgres-pvc
  namespace: lokstra-example
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 10Gi
```

### 2. Deploy to Kubernetes
```bash
kubectl apply -f namespace.yaml
kubectl apply -f configmap.yaml
kubectl apply -f postgres.yaml
kubectl apply -f deployment.yaml
kubectl apply -f service.yaml
```

## Method 4: Cloud Deployment (AWS/GCP/Azure)

### AWS with ECS

1. **Create ECR Repository:**
```bash
aws ecr create-repository --repository-name lokstra-example
```

2. **Build and Push Docker Image:**
```bash
docker build -t lokstra-example .
docker tag lokstra-example:latest <account-id>.dkr.ecr.<region>.amazonaws.com/lokstra-example:latest
docker push <account-id>.dkr.ecr.<region>.amazonaws.com/lokstra-example:latest
```

3. **Create ECS Task Definition and Service using AWS CLI or Console**

### GCP with Cloud Run

1. **Build and Deploy:**
```bash
gcloud builds submit --tag gcr.io/PROJECT_ID/lokstra-example
gcloud run deploy --image gcr.io/PROJECT_ID/lokstra-example --platform managed
```

## Environment-Specific Configuration

### Development
```yaml
server:
  global_setting:
    log_level: "debug"

services:
  - name: "db_pool"
    config:
      host: "localhost"
      max_connections: 10
```

### Staging
```yaml
server:
  global_setting:
    log_level: "info"

services:
  - name: "db_pool"
    config:
      host: "staging-db.example.com"
      max_connections: 20
```

### Production
```yaml
server:
  global_setting:
    log_level: "warn"

services:
  - name: "db_pool"
    config:
      host: "prod-db.example.com"
      max_connections: 50
      ssl_mode: "require"
```

## Health Checks and Monitoring

### Health Check Endpoint
The application provides a health check at `/api/health`:
```bash
curl http://localhost:8080/api/health
```

### Kubernetes Probes
Add to your deployment:
```yaml
livenessProbe:
  httpGet:
    path: /api/health
    port: 8080
  initialDelaySeconds: 30
  periodSeconds: 10

readinessProbe:
  httpGet:
    path: /api/health
    port: 8080
  initialDelaySeconds: 5
  periodSeconds: 5
```

## Scaling Considerations

1. **Database Connection Pooling**: Adjust `max_connections` based on your database limits
2. **Horizontal Scaling**: The application is stateless and can be scaled horizontally
3. **Database**: Consider read replicas for read-heavy workloads
4. **Caching**: Add Redis caching for frequently accessed data

## Security Best Practices

1. **Environment Variables**: Use environment variables for sensitive data
2. **Database SSL**: Enable SSL for database connections in production
3. **HTTPS**: Use HTTPS in production deployments
4. **Network Policies**: Implement Kubernetes network policies for security
5. **Secrets Management**: Use proper secrets management (Kubernetes secrets, AWS Secrets Manager, etc.)

## Backup and Recovery

### Database Backup
```bash
# Automated backup script
pg_dump -h $DB_HOST -U $DB_USER $DB_NAME > backup_$(date +%Y%m%d_%H%M%S).sql
```

### Kubernetes Backup
Use tools like Velero for Kubernetes cluster backups.

## Troubleshooting

### Common Issues

1. **Database Connection Issues:**
   - Check database credentials
   - Verify network connectivity
   - Check database service status

2. **Module Loading Issues:**
   - Ensure module path is correct
   - Verify module compilation
   - Check required services availability

3. **Performance Issues:**
   - Monitor database connection pool
   - Check query performance
   - Monitor application logs
