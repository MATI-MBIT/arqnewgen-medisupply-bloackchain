# Alchemy WebSocket Micro Service

Microservicio en Go para conexiones WebSocket con Alchemy Ethereum API.

## Descripción

Este servicio proporciona conectividad WebSocket con la API de Alchemy para interactuar con la blockchain de Ethereum Sepolia. Está diseñado para manejar conexiones en tiempo real y eventos de blockchain.

## Variables de Entorno

El servicio requiere las siguientes variables de entorno:

- `ALCHEMY_API_KEY` (requerida): API key de Alchemy
- `ALCHEMY_WS_URL` (opcional): URL del WebSocket de Alchemy (default: "wss://eth-sepolia.g.alchemy.com/v2")
- `PORT` (opcional): Puerto del servidor HTTP (default: "8081")

## Configuración de Secretos

Este servicio **reutiliza** los secretos de `blockchain-secrets` que también usa `crear-lote-micro`. El secreto debe contener:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: blockchain-secrets
type: Opaque
data:
  sepolia-rpc: <base64-encoded-rpc-endpoint>
  sepolia-ws: <base64-encoded-ws-endpoint>
  alchemy-api-key: <base64-encoded-alchemy-api-key>
```

### Crear Secretos

#### Opción 1: Script Interactivo (Recomendado)
```bash
# Desde el directorio k8s/
./scripts/create-blockchain-secrets.sh [namespace]
```

#### Opción 2: Makefile
```bash
# Desde el directorio k8s/
make create-blockchain-secrets NAMESPACE=your-namespace
```

#### Opción 3: Manual
```bash
# Create the secret manually
kubectl create secret generic blockchain-secrets \
  --from-literal=sepolia-rpc="https://eth-sepolia.g.alchemy.com/v2/YOUR_API_KEY" \
  --from-literal=sepolia-ws="wss://eth-sepolia.g.alchemy.com/v2/YOUR_API_KEY" \
  --from-literal=alchemy-api-key="YOUR_ALCHEMY_API_KEY" \
  -n your-namespace
```

## Despliegue

### Prerequisitos
1. Cluster de Kubernetes funcionando
2. Helm 3.x instalado
3. Secretos de blockchain creados (ver sección anterior)

### Instalación

```bash
# Desde el directorio k8s/
helm install alchemy-websocket-micro ./microservice \
  -f config/services/alchemy-websocket-micro/alchemy-websocket-micro-service-values.yaml \
  -n your-namespace
```

### Actualización

```bash
# Desde el directorio k8s/
helm upgrade alchemy-websocket-micro ./microservice \
  -f config/services/alchemy-websocket-micro/alchemy-websocket-micro-service-values.yaml \
  -n your-namespace
```

### Desinstalación

```bash
helm uninstall alchemy-websocket-micro -n your-namespace
```

## Configuración del Servicio

### Recursos
- **CPU**: 100m (request) / 200m (limit)
- **Memory**: 128Mi (request) / 256Mi (limit)

### Puertos
- **Service Port**: 8081
- **Target Port**: 8081

### Health Checks
- **Liveness Probe**: `/api/v1/health` (puerto 8081)
- **Readiness Probe**: `/api/v1/health` (puerto 8081)

### Istio Service Mesh
- Inyección de sidecar habilitada
- Configurado para trabajar con Istio

## Monitoreo

### Prometheus Metrics
- **Endpoint**: `/metrics`
- **Puerto**: 9090
- **Habilitado**: Sí

### Logs
```bash
# Ver logs del servicio
kubectl logs -f deployment/alchemy-websocket-micro -n your-namespace

# Ver logs con Istio sidecar
kubectl logs -f deployment/alchemy-websocket-micro -c alchemy-websocket-micro -n your-namespace
```

## Troubleshooting

### Verificar Secretos
```bash
# Verificar que el secreto existe
kubectl get secret blockchain-secrets -n your-namespace

# Ver las claves del secreto (sin valores)
kubectl get secret blockchain-secrets -n your-namespace -o jsonpath='{.data}' | jq 'keys'
```

### Verificar Configuración
```bash
# Ver variables de entorno del pod
kubectl exec -it deployment/alchemy-websocket-micro -n your-namespace -- env | grep -E "(ALCHEMY|PORT)"
```

### Verificar Conectividad
```bash
# Test health endpoint
kubectl port-forward svc/alchemy-websocket-micro 8081:8081 -n your-namespace
curl http://localhost:8081/api/v1/health
```

## Arquitectura

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   Client App    │───▶│ Alchemy WS Micro │───▶│  Alchemy API    │
└─────────────────┘    └──────────────────┘    └─────────────────┘
                              │
                              ▼
                       ┌──────────────────┐
                       │ Ethereum Sepolia │
                       └──────────────────┘
```

## Notas Importantes

1. **Secretos Compartidos**: Este servicio comparte los secretos `blockchain-secrets` con `crear-lote-micro`
2. **Red Sepolia**: Configurado para trabajar con Ethereum Sepolia testnet
3. **WebSocket**: Mantiene conexiones persistentes con Alchemy WebSocket API
4. **Istio Ready**: Configurado para trabajar con Istio Service Mesh
5. **Prometheus**: Métricas expuestas para monitoreo