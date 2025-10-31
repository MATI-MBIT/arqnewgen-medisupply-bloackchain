# AlchemyWebSocketMicro

Microservicio WebSocket en Go para monitoreo en tiempo real de transacciones Ethereum usando Alchemy WebSocket API.

## 🚀 Características

- **WebSocket Server**: Conexiones en tiempo real para múltiples clientes
- **Alchemy Integration**: Conexión directa con Alchemy WebSocket API
- **Multi-Contract Monitoring**: Monitoreo simultáneo de múltiples contratos
- **Auto-Reconnection**: Reconexión automática en caso de desconexión
- **Detailed Logging**: Logging completo de todas las operaciones WebSocket
- **REST API**: Endpoints complementarios para gestión y estado

## 📡 Endpoints

### WebSocket
- `ws://localhost:8081/ws/monitor/{contractAddress}` - Monitoreo en tiempo real

### REST API
- `GET /api/v1/health` - Health check
- `GET /api/v1/monitor/status` - Estado de suscripciones activas
- `POST /api/v1/monitor/start/{contractAddress}` - Información de inicio
- `POST /api/v1/monitor/stop/{contractAddress}` - Información de parada
- `GET /` - Información del servicio

## 🔧 Configuración

### Variables de Entorno

```bash
ALCHEMY_API_KEY=your_alchemy_api_key_here
ALCHEMY_WS_URL=wss://eth-sepolia.g.alchemy.com/v2
PORT=8081
```

### Configuración Local

1. Copiar `.env.example` a `.env`
2. Configurar tu Alchemy API Key
3. Ejecutar el microservicio

## 🏃‍♂️ Ejecución

### Desarrollo Local

```bash
# Instalar dependencias
go mod tidy

# Ejecutar
go run main.go
```

### Docker

```bash
# Build y run
docker-compose up --build

# Solo run (si ya está built)
docker-compose up
```

### Docker manual

```bash
# Build
docker build -t alchemy-websocket-micro .

# Run
docker run -p 8081:8081 \
  -e ALCHEMY_API_KEY=your_key_here \
  alchemy-websocket-micro
```

## 📝 Uso

### Conectar via WebSocket

```bash
# Usando wscat
wscat -c ws://localhost:8081/ws/monitor/0x9d70c560cE7D6EDAaf4562E980136D21Fd0fbdc9

# Usando curl para health check
curl http://localhost:8081/api/v1/health

# Ver estado de monitoreo
curl http://localhost:8081/api/v1/monitor/status
```

### Ejemplo con JavaScript

```javascript
const ws = new WebSocket('ws://localhost:8081/ws/monitor/0x9d70c560cE7D6EDAaf4562E980136D21Fd0fbdc9');

ws.onopen = function() {
    console.log('Conectado al monitoreo de transacciones');
};

ws.onmessage = function(event) {
    const data = JSON.parse(event.data);
    console.log('Nueva transacción:', data);
};

ws.onclose = function() {
    console.log('Conexión cerrada');
};
```

## 📊 Formato de Mensajes

### Mensaje de Conexión
```json
{
  "type": "connected",
  "contractAddress": "0x9d70c560cE7D6EDAaf4562E980136D21Fd0fbdc9",
  "data": {
    "status": "monitoring started"
  },
  "timestamp": 1640995200
}
```

### Mensaje de Transacción
```json
{
  "type": "transaction",
  "contractAddress": "0x9d70c560cE7D6EDAaf4562E980136D21Fd0fbdc9",
  "data": {
    "hash": "0x...",
    "from": "0x...",
    "to": "0x...",
    "value": "0x0",
    "gas": "0x5208",
    "gasPrice": "0x...",
    "input": "0x...",
    "blockNumber": "0x...",
    "blockHash": "0x...",
    "transactionIndex": "0x..."
  },
  "timestamp": 1640995200
}
```

### Mensaje de Error
```json
{
  "type": "error",
  "contractAddress": "0x9d70c560cE7D6EDAaf4562E980136D21Fd0fbdc9",
  "data": {
    "error": "Connection failed"
  },
  "timestamp": 1640995200
}
```

## 🔍 Logging

El servicio proporciona logging detallado de todas las operaciones:

- 🚀 Inicio del servicio
- 🔌 Conexiones WebSocket entrantes y salientes
- 📡 Comunicación con Alchemy WebSocket
- 📨 Mensajes recibidos y enviados
- 🔄 Reconexiones automáticas
- ❌ Errores y excepciones

### Ejemplo de Logs

```
🚀 Iniciando AlchemyWebSocketMicro...
⚙️ Configuración cargada - Puerto: 8081
🔗 Alchemy WebSocket URL: wss://eth-sepolia.g.alchemy.com/v2
🔌 Iniciando conexión con Alchemy...
✅ Conectado exitosamente a Alchemy WebSocket
✅ Servicio de Alchemy iniciado exitosamente
🌐 AlchemyWebSocketMicro iniciando en puerto 8081
🔌 Nueva conexión WebSocket solicitada para contrato: 0x9d70c560cE7D6EDAaf4562E980136D21Fd0fbdc9
✅ WebSocket connection establecida para: 0x9d70c560cE7D6EDAaf4562E980136D21Fd0fbdc9
📡 Enviando suscripción a Alchemy para: 0x9d70c560cE7D6EDAaf4562E980136D21Fd0fbdc9
📤 Enviando request a Alchemy: {"jsonrpc":"2.0","method":"eth_subscribe",...}
✅ Request enviado exitosamente a Alchemy
📨 Mensaje recibido de Alchemy: {"jsonrpc":"2.0","id":1640995200,"result":"0x..."}
✅ Suscripción creada exitosamente - ID: 0x...
🔔 Notificación de transacción recibida - Subscription: 0x...
💰 Procesando transacción - Subscription: 0x...
📤 Enviando a 1 clientes conectados
✅ Mensaje enviado exitosamente a cliente
```

## 🏗️ Arquitectura

```
Cliente WebSocket ←→ AlchemyWebSocketMicro ←→ Alchemy WebSocket API
                            ↓
                    Logging & Processing
                            ↓
                    Broadcast a múltiples clientes
```

## 🔒 Seguridad

- Validación de direcciones de contratos Ethereum
- CORS configurado para desarrollo
- Usuario no-root en Docker
- Health checks integrados

## 🐛 Troubleshooting

### Error de conexión a Alchemy
- Verificar que `ALCHEMY_API_KEY` esté configurada correctamente
- Verificar conectividad a internet
- Revisar logs para detalles específicos

### WebSocket no conecta
- Verificar que el puerto 8081 esté disponible
- Verificar formato de dirección de contrato (42 caracteres, inicia con 0x)
- Revisar logs del servidor

### Sin transacciones recibidas
- Verificar que el contrato tenga actividad en Sepolia
- Verificar que la dirección del contrato sea correcta
- Revisar logs de Alchemy para errores de suscripción

## 📦 Dependencias

- `github.com/gin-gonic/gin` - Framework web
- `github.com/gorilla/websocket` - WebSocket support
- `github.com/joho/godotenv` - Environment variables

## 🤝 Contribución

1. Fork el proyecto
2. Crear feature branch
3. Commit cambios
4. Push al branch
5. Crear Pull Request