# Kubernetes Deployment Guide

This directory contains Kubernetes manifests for deploying Nostos to a Kubernetes cluster.

## Prerequisites

- Kubernetes cluster (1.24+)
- `kubectl` configured to access your cluster
- Container registry access (for pushing the image)
- Kafka cluster running (either in-cluster or external)

## Files

- **configmap.yaml** - Application configuration (Kafka brokers, topics, etc.)
- **deployment.yaml** - Nostos ingress service deployment
- **service.yaml** - ClusterIP and LoadBalancer services
- **rbac.yaml** - ServiceAccount and RBAC setup

## Building and Pushing the Image

```bash
# Build the Docker image
docker build -t your-registry/nostos-ingress:latest go_ingress/

# Push to your registry
docker push your-registry/nostos-ingress:latest
```

## Deploying to Kubernetes

### 1. Update the image reference
Edit `deployment.yaml` and replace `your-registry/nostos-ingress:latest` with your actual image:

```bash
sed -i 's|your-registry/nostos-ingress:latest|your-actual-image|g' k8s/deployment.yaml
```

### 2. Configure Kafka brokers
Edit `configmap.yaml` and update `KAFKA_BROKERS` to point to your Kafka cluster:

```yaml
KAFKA_BROKERS: kafka-0.kafka-headless.kafka:9092,kafka-1.kafka-headless.kafka:9092,kafka-2.kafka-headless.kafka:9092
```

### 3. Deploy the manifests

```bash
# Deploy all manifests
kubectl apply -f k8s/

# Or deploy individually
kubectl apply -f k8s/configmap.yaml
kubectl apply -f k8s/rbac.yaml
kubectl apply -f k8s/deployment.yaml
kubectl apply -f k8s/service.yaml
```

### 4. Verify deployment

```bash
# Check deployment status
kubectl get deployments -l app=nostos

# Check pods
kubectl get pods -l app=nostos

# View logs
kubectl logs -l app=nostos -f

# Check services
kubectl get svc nostos-ingress-lb
```

## Accessing the Service

### Internal (ClusterIP)
```bash
# From within the cluster
curl http://nostos-ingress/ingest
```

### External (LoadBalancer)
```bash
# Get the external IP
kubectl get svc nostos-ingress-lb

# Make a request
curl -X POST http://<EXTERNAL-IP>/ingest \
  -H "Content-Type: application/json" \
  -d '{"your": "data"}'
```

## Configuration

Environment variables can be customized in `configmap.yaml`:

- `KAFKA_BROKERS` - Comma-separated list of Kafka brokers
- `KAFKA_TOPIC` - Kafka topic to publish events to
- `SERVER_ADDR` - Server bind address (default: `:8080`)

## Scaling

```bash
# Scale to N replicas
kubectl scale deployment nostos-ingress --replicas=5

# Or edit the deployment
kubectl edit deployment nostos-ingress
```

## Rolling Updates

The deployment uses a rolling update strategy. To update the image:

```bash
# Update image
kubectl set image deployment/nostos-ingress \
  ingress=your-registry/nostos-ingress:v1.2.3

# Watch the rollout
kubectl rollout status deployment/nostos-ingress

# Rollback if needed
kubectl rollout undo deployment/nostos-ingress
```

## Health Checks

The deployment includes:
- **Liveness probe** - Restarts unhealthy containers
- **Readiness probe** - Removes unhealthy pods from load balancing

Ensure your application implements these endpoints:
- `GET /healthz` - Liveness check
- `GET /readyz` - Readiness check

## Security

The deployment enforces:
- Non-root user (UID 10001)
- Read-only root filesystem
- No privilege escalation
- Dropped Linux capabilities

## Pod Disruption Budget (Optional)

For production, consider adding a PDB to ensure availability during cluster updates:

```yaml
apiVersion: policy/v1
kind: PodDisruptionBudget
metadata:
  name: nostos-ingress-pdb
spec:
  minAvailable: 2
  selector:
    matchLabels:
      app: nostos
      component: ingress
```

## Monitoring (Future)

Once metrics are implemented, add Prometheus ServiceMonitor:

```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: nostos-ingress
spec:
  selector:
    matchLabels:
      app: nostos
  endpoints:
    - port: http
      path: /metrics
```
