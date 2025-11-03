# MediSupply Blockchain - Trazabilidad de Lotes Farmacéuticos

Sistema blockchain completo para la trazabilidad de lotes de productos farmacéuticos con enfoque en la integridad de la cadena de frío y transferencia de custodia. Desarrollado por **Grupo 2 - ArqNewGen - MATI** como prueba de concepto para MediSupply.

## 🏗️ Arquitectura del Sistema

### Componentes Principales

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────────┐
│   Smart Contract│    │   Microservicios │    │   Infraestructura   │
│   LoteTracing   │◄──►│   CrearLoteMicro │◄──►│   Kubernetes + EDA  │
│   (Sepolia)     │    │   AlchemyWS      │    │   Kafka + MQTT      │
└─────────────────┘    └──────────────────┘    └─────────────────────┘
```

### Flujo Event-Driven Architecture (EDA)

```
IoT Sensors → MQTT → EMQX → Microservicios → Kafka → Blockchain → Dashboard
```

## 🚀 Características Principales

### 🔗 Smart Contract (Ethereum Sepolia)
- **Trazabilidad completa** de lotes farmacéuticos
- **Monitoreo de cadena de frío** con alertas automáticas
- **Transferencia de custodia** entre entidades
- **Eventos inmutables** para auditoría completa
- **Gestión de múltiples lotes** en un mismo contrato

### 🛠️ Microservicios
- **CrearLoteMicro**: API REST para interacción con blockchain
- **AlchemyWebSocketMicro**: Monitoreo en tiempo real de eventos
- **Arquitectura escalable** con Docker y Kubernetes

### 🏢 Infraestructura Cloud-Native
- **Kubernetes** con Istio Service Mesh
- **Apache Kafka** para procesamiento de eventos
- **EMQX MQTT** para dispositivos IoT
- **KEDA** para autoescalado basado en eventos
- **Observabilidad completa** con Prometheus, Grafana y Jaeger

## 📁 Estructura del Proyecto

```
arqnewgen-medisupply-blockchain/
├── smartcontract/lotetracing/          # Smart Contract Ethereum
│   ├── contracts/LoteTracing.sol       # Contrato principal
│   ├── test/                          # Tests automatizados
│   ├── scripts/                       # Scripts de despliegue
│   └── hardhat.config.ts              # Configuración Hardhat
├── services/                          # Microservicios
│   ├── CrearLoteMicro/               # API REST principal
│   ├── AlchemyWebSocketMicro/        # WebSocket para eventos
│   ├── mqtt-event-generator/         # Generador de eventos IoT
│   └── mqtt-order-event-client/      # Cliente de eventos
├── k8s/                              # Infraestructura Kubernetes
│   ├── istio/                        # Service Mesh
│   ├── kafka/                        # Apache Kafka
│   ├── mqtt/                         # EMQX MQTT Broker
│   └── microservice/                 # Charts de microservicios
└── README.md                         # Este archivo
```

## 🎯 Casos de Uso

### 1. Trazabilidad de Medicamentos
- Registro de lotes desde fabricación hasta paciente final
- Monitoreo continuo de temperatura durante transporte
- Alertas automáticas por ruptura de cadena de frío
- Auditoría completa e inmutable

### 2. Transferencia de Custodia
- Cambio de propietario entre fabricante → distribuidor → farmacia
- Registro automático de transferencias en blockchain
- Verificación de autenticidad en cada paso

### 3. Monitoreo IoT en Tiempo Real
- Sensores de temperatura conectados via MQTT
- Procesamiento de eventos en tiempo real
- Escalado automático basado en carga de eventos

## 🚀 Inicio Rápido

### Prerrequisitos

- **Docker** y **Docker Compose**
- **kubectl** y **helm**
- **Node.js 18+** (para smart contracts)
- **Go 1.21+** (para microservicios)
- **Kind** o **Minikube** (para desarrollo local)

### 1. Clonar el Repositorio

```bash
git clone https://github.com/tu-org/arqnewgen-medisupply-blockchain.git
cd arqnewgen-medisupply-blockchain
```

### 2. Desplegar Infraestructura Kubernetes

```bash
cd k8s
make init     # Crear cluster local
make deploy   # Desplegar toda la infraestructura
make status   # Verificar estado
```

### 3. Compilar y Desplegar Smart Contract

```bash
cd smartcontract/lotetracing
npm install
npx hardhat compile
npx hardhat run scripts/demo-lote-tracing.ts --network sepolia
```

### 4. Ejecutar Microservicios

```bash
# CrearLoteMicro
cd services/CrearLoteMicro
go mod tidy
go run main.go

# En otra terminal - AlchemyWebSocketMicro
cd services/AlchemyWebSocketMicro
go mod tidy
go run main.go
```

### 5. Acceder a Dashboards

| Servicio | URL | Credenciales |
|----------|-----|--------------|
| Kafka UI | http://localhost:9090 | - |
| EMQX Dashboard | http://localhost:18083 | admin/public |
| RabbitMQ Management | http://localhost:15672 | guest/guest |
| Kiali (Istio) | http://localhost:20001 | - |

## 📊 API Endpoints

### CrearLoteMicro (Puerto 8080)

| Endpoint | Método | Descripción |
|----------|--------|-------------|
| `/api/v1/health` | GET | Health check |
| `/api/v1/lote/crear` | POST | Crear nuevo lote (deploy contrato) |
| `/api/v1/lote/nuevo` | POST | Crear lote en contrato existente |
| `/api/v1/lote/temperatura` | POST | Registrar temperatura |
| `/api/v1/lote/transferir` | POST | Transferir custodia |
| `/api/v1/lote/info/{address}` | GET | Obtener información del lote |
| `/api/v1/lote/cadena/{address}` | GET | Historial blockchain completo |
| `/api/v1/debug/contrato/{address}` | GET | Diagnóstico de contrato |

### Ejemplo: Crear Nuevo Lote

```bash
curl -X POST http://localhost:8080/api/v1/lote/crear \
  -H "Content-Type: application/json" \
  -d '{
    "loteId": "LOTE_MEDICAMENTO_001",
    "temperaturaMin": 2,
    "temperaturaMax": 8,
    "walletAddress": "0x742d35Cc6634C0532925a3b8D4C9db96590c6C87",
    "privateKey": "0x..."
  }'
```

## 🔧 Configuración

### Variables de Entorno

#### CrearLoteMicro
```bash
SEPOLIA_RPC=https://sepolia.infura.io/v3/YOUR_PROJECT_ID
SEPOLIA_WS=wss://eth-sepolia.g.alchemy.com/v2/YOUR_PROJECT_ID
CHAIN_ID=11155111
PORT=8080
```

#### Kubernetes
```bash
# Configurar provider de cluster
export K8S_PROVIDER=kind  # o minikube

# Configurar registry de imágenes
export DOCKER_REGISTRY=your-registry.com
```

## 🧪 Testing

### Smart Contract Tests

```bash
cd smartcontract/lotetracing
npm test
```

### Microservicios Tests

```bash
cd services/CrearLoteMicro
go test ./...
```

### Tests de Integración

```bash
# Usar colección de Postman incluida
cd services/CrearLoteMicro
# Importar CrearLoteMicro.postman_collection.json en Postman
```

## 📈 Monitoreo y Observabilidad

### Métricas Disponibles
- **Transacciones blockchain** por segundo
- **Eventos IoT procesados** por minuto
- **Latencia de APIs** y tiempo de respuesta
- **Estado de contratos** y gas utilizado

### Logs Centralizados
- **Microservicios**: Logs estructurados en JSON
- **Kubernetes**: Agregación con Fluentd/Fluent Bit
- **Blockchain**: Eventos y transacciones trackeadas

### Alertas
- **Ruptura de cadena de frío** → Slack/Email
- **Fallos de transacciones** → PagerDuty
- **Sobrecarga de sistema** → Autoescalado KEDA

## 🔒 Seguridad

### Smart Contract
- **Auditoría de código** con herramientas estáticas
- **Tests de penetración** automatizados
- **Gestión de claves** con HSM en producción

### Microservicios
- **Autenticación JWT** (en desarrollo)
- **Rate limiting** por IP y usuario
- **Validación de entrada** estricta

### Infraestructura
- **Istio mTLS** para comunicación entre servicios
- **Network policies** de Kubernetes
- **Secrets management** con Vault

## 🚀 Roadmap

### Fase 1 - MVP ✅
- [x] Smart contract básico
- [x] API REST funcional
- [x] Infraestructura Kubernetes
- [x] Monitoreo básico

### Fase 2 - Producción 🔄
- [ ] Autenticación y autorización
- [ ] Dashboard web completo
- [ ] Integración con ERPs
- [ ] Alertas avanzadas

### Fase 3 - Escalabilidad 📋
- [ ] Multi-chain support
- [ ] IA para predicción de fallos
- [ ] Integración con reguladores
- [ ] Mobile app

## 🤝 Contribución

### Desarrollo Local

1. **Fork** el repositorio
2. **Crear branch** para feature: `git checkout -b feature/nueva-funcionalidad`
3. **Commit** cambios: `git commit -am 'Agregar nueva funcionalidad'`
4. **Push** al branch: `git push origin feature/nueva-funcionalidad`
5. **Crear Pull Request**

### Estándares de Código

- **Go**: `gofmt` y `golint`
- **Solidity**: `prettier-plugin-solidity`
- **TypeScript**: `eslint` y `prettier`
- **Commits**: Conventional Commits

### Testing

- **Cobertura mínima**: 80%
- **Tests unitarios** obligatorios
- **Tests de integración** para APIs
- **Tests de contrato** con Hardhat

## 📄 Licencia

MIT License - Ver [LICENSE](LICENSE) para más detalles.

## 👥 Equipo

**Grupo 2 - ArqNewGen - MATI**

- **Blockchain Development**: Smart contracts y integración Web3
- **Backend Development**: Microservicios y APIs REST
- **DevOps Engineering**: Kubernetes e infraestructura cloud
- **IoT Integration**: Sensores y protocolos MQTT

## 📞 Soporte

- **Issues**: [GitHub Issues](https://github.com/tu-org/arqnewgen-medisupply-blockchain/issues)
- **Documentación**: [Wiki del proyecto](https://github.com/tu-org/arqnewgen-medisupply-blockchain/wiki)
- **Email**: medisupply-blockchain@mati.edu

---

**🏥 MediSupply Blockchain - Revolucionando la trazabilidad farmacéutica con tecnología blockchain** 🚀
