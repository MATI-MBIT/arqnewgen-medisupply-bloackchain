# Crear Lote Micro Service - Kubernetes Deployment

This directory contains the Kubernetes configuration for the `crear-lote-micro` service, which provides blockchain integration with Ethereum Sepolia network.

## Prerequisites

1. Kubernetes cluster with Helm installed
2. Blockchain endpoints (Alchemy or Infura)
3. kubectl configured to access your cluster

## Integrated Make Commands

The easiest way to deploy the service is using the integrated make commands that handle secret creation automatically:

### Quick Deployment

```bash
# Deploy with automatic secret checking (will prompt for secrets if needed)
make deploy-service SERVICE=crear-lote-micro NAMESPACE=default

# Build and deploy in one command
make build-deploy-service SERVICE=crear-lote-micro NAMESPACE=default
```

### Interactive Secret Creation

```bash
# Create secrets interactively for any service
make create-secrets-interactive

# Create blockchain secrets specifically
make create-blockchain-secrets NAMESPACE=default
```

### How It Works

1. **Automatic Detection**: The make commands automatically detect that `crear-lote-micro` requires blockchain secrets
2. **Secret Validation**: Before deployment, it checks if the required `blockchain-secrets` secret exists
3. **Interactive Prompts**: If secrets don't exist, it provides clear instructions on how to create them
4. **Seamless Deployment**: Once secrets are available, deployment proceeds automatically

### Example Workflow

```bash
# 1. Try to deploy (will detect missing secrets)
make deploy-service SERVICE=crear-lote-micro NAMESPACE=medisupply

# 2. If secrets are missing, create them interactively
./k8s/scripts/create-blockchain-secrets.sh medisupply

# 3. Deploy again (will succeed now)
make deploy-service SERVICE=crear-lote-micro NAMESPACE=medisupply
```

This integration ensures that:
- Secrets are never hardcoded in configuration files
- Deployment fails fast if required secrets are missing
- The process is consistent across all environments
- Security best practices are enforced automatically

## Secret Management

The service requires blockchain endpoints to be stored as Kubernetes secrets for security.

### Option 1: Using the provided script (Recommended)

```bash
# Create secrets interactively
./k8s/scripts/create-blockchain-secrets.sh [namespace]

# Example for default namespace
./k8s/scripts/create-blockchain-secrets.sh

# Example for specific namespace
./k8s/scripts/create-blockchain-secrets.sh mediorder
```

### Option 2: Manual secret creation

```bash
# Create the secret manually
kubectl create secret generic blockchain-secrets \
  --from-literal=sepolia-rpc="https://eth-sepolia.g.alchemy.com/v2/YOUR_API_KEY" \
  --from-literal=sepolia-ws="wss://eth-sepolia.g.alchemy.com/v2/YOUR_API_KEY" \
  -n your-namespace
```

### Option 3: Using YAML file

Create a secret YAML file (DO NOT commit this to git):

```yaml
# blockchain-secrets.yaml (DO NOT COMMIT)
apiVersion: v1
kind: Secret
metadata:
  name: blockchain-secrets
  namespace: your-namespace
type: Opaque
data:
  sepolia-rpc: <base64-encoded-rpc-endpoint>
  sepolia-ws: <base64-encoded-ws-endpoint>
```

Apply it:
```bash
kubectl apply -f blockchain-secrets.yaml
```

## Deployment

Once the secrets are created, deploy the service:

```bash
# Install the service
helm install crear-lote-micro ./k8s/microservice \
  -f k8s/config/services/crear-lote-micro/crear-lote-micro-service-values.yaml \
  -n your-namespace

# Upgrade the service
helm upgrade crear-lote-micro ./k8s/microservice \
  -f k8s/config/services/crear-lote-micro/crear-lote-micro-service-values.yaml \
  -n your-namespace
```

## Verification

### Check deployment status
```bash
kubectl get pods -l app=crear-lote-micro -n your-namespace
kubectl get svc -l app=crear-lote-micro -n your-namespace
```

### Check logs
```bash
kubectl logs -l app=crear-lote-micro -n your-namespace -f
```

### Test health endpoint
```bash
# Port forward to test locally
kubectl port-forward svc/crear-lote-micro 8080:8080 -n your-namespace

# Test health endpoint
curl http://localhost:8080/api/v1/health
```

### Verify blockchain connection
```bash
# Test blockchain connection
curl http://localhost:8080/api/v1/debug/conexion
```

## Environment Variables

The service uses the following environment variables:

| Variable | Source | Description |
|----------|--------|-------------|
| `SEPOLIA_RPC` | Secret | Ethereum Sepolia RPC endpoint |
| `SEPOLIA_WS` | Secret | Ethereum Sepolia WebSocket endpoint |
| `PORT` | ConfigMap | HTTP server port (default: 8080) |
| `POD_NAME` | Kubernetes | Pod name for logging |
| `POD_NAMESPACE` | Kubernetes | Pod namespace for logging |

## Security Considerations

1. **Never commit secrets to git** - Use the secret management options above
2. **Rotate API keys regularly** - Update secrets when rotating blockchain provider keys
3. **Use RBAC** - Ensure proper access controls for secret management
4. **Monitor usage** - Keep track of blockchain API usage and costs

## Troubleshooting

### Secret not found error
```bash
# Check if secret exists
kubectl get secret blockchain-secrets -n your-namespace

# Check secret contents (keys only)
kubectl get secret blockchain-secrets -n your-namespace -o jsonpath='{.data}' | jq 'keys'
```

### Pod not starting
```bash
# Check pod events
kubectl describe pod -l app=crear-lote-micro -n your-namespace

# Check logs
kubectl logs -l app=crear-lote-micro -n your-namespace
```

### Blockchain connection issues
1. Verify the RPC endpoints are correct
2. Check API key validity and rate limits
3. Ensure network connectivity from the cluster

## Configuration Options

The `crear-lote-micro-service-values.yaml` file supports the following configurations:

- **Resources**: CPU/Memory limits and requests
- **Replicas**: Number of pod replicas
- **Health checks**: Liveness and readiness probe settings
- **Istio**: Service mesh integration
- **Monitoring**: Prometheus metrics configuration

Modify the values file as needed for your environment.
