# KEDA Scaling Configuration - MediSupply EDA

Esta carpeta contiene las configuraciones de autoescalado basado en eventos para los servicios de MediSupply usando KEDA.

## 📋 Servicios Configurados

### 1. mqtt-order-event-client
- **Archivo**: `mqtt-order-event-client-scaler.yaml`
- **Namespace**: `medisupply`
- **Trigger**: Kafka topic `events-order`
- **Escalado**: 1-10 pods basado en lag de mensajes (threshold: 10)

## 🚀 Instalación

### Prerequisitos

1. KEDA debe estar instalado en el cluster:
   ```bash
   helm install keda ../keda --namespace keda-system --create-namespace
   ```

2. El namespace `medisupply` debe existir:
   ```bash
   kubectl create namespace medisupply
   ```

### Aplicar Configuraciones

```bash
# Aplicar la configuración de escalado
kubectl apply -f k8s/keda-scaling/mqtt-order-event-client-scaler.yaml

# O aplicar todo el directorio
kubectl apply -f k8s/keda-scaling/
```

## 📊 Monitoreo

### Verificar ScaledObjects
```bash
# Ver todos los ScaledObjects
kubectl get scaledobjects -A

# Ver detalles de un ScaledObject específico
kubectl describe scaledobject mqtt-order-event-client-scaler -n medisupply
```

### Verificar HPA Generados
```bash
# KEDA crea automáticamente HPAs
kubectl get hpa -A

# Ver métricas de escalado
kubectl get --raw "/apis/external.metrics.k8s.io/v1beta1"
```

### Logs de KEDA
```bash
# Ver logs del operador KEDA
kubectl logs deployment/keda-operator -n keda-system

# Ver logs del metrics server
kubectl logs deployment/keda-operator-metrics-apiserver -n keda-system
```

## ⚙️ Configuración Personalizada

### Parámetros Comunes

- **minReplicaCount**: Número mínimo de pods
- **maxReplicaCount**: Número máximo de pods
- **pollingInterval**: Frecuencia de verificación (segundos)
- **cooldownPeriod**: Tiempo de espera antes de escalar hacia abajo (segundos)

### Kafka Scaler Parameters

- **bootstrapServers**: Dirección del cluster Kafka
- **consumerGroup**: Grupo de consumidores
- **topic**: Topic a monitorear
- **lagThreshold**: Umbral de lag para disparar escalado
- **offsetResetPolicy**: Política de reset (latest/earliest)

### RabbitMQ Scaler Parameters

- **host**: URL de conexión a RabbitMQ
- **queueName**: Nombre de la cola
- **queueLength**: Longitud de cola para disparar escalado
- **vhostName**: Virtual host (por defecto '/')

## 🔧 Troubleshooting

### Problemas Comunes

1. **ScaledObject no escala**:
   ```bash
   kubectl describe scaledobject <name> -n <namespace>
   kubectl logs deployment/keda-operator -n keda-system
   ```

2. **Métricas no disponibles**:
   ```bash
   kubectl get --raw "/apis/external.metrics.k8s.io/v1beta1"
   ```

3. **Problemas de autenticación**:
   ```bash
   kubectl get secrets -n <namespace>
   kubectl describe triggerauthentication <name> -n <namespace>
   ```

## 📈 Métricas y Alertas

Las métricas de KEDA están disponibles en Prometheus:
- `keda_metrics_adapter_scaler_metrics_value`
- `keda_metrics_adapter_scaler_errors`
- `keda_metrics_adapter_scaled_object_errors`

Configura alertas para:
- Errores de escalado
- Métricas no disponibles
- Escalado excesivo