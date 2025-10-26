# CrearLoteMicro

Microservicio en Go para interactuar con el contrato LoteTracing en la red Sepolia de Ethereum.

## Funcionalidades

- **Health Check**: Verifica el estado del microservicio
- **Debug Conexión**: Verifica la conexión a Sepolia y obtiene información de la blockchain
- **Crear Lote**: Despliega un nuevo contrato LoteTracing
- **Registrar Temperatura**: Registra lecturas de temperatura en un lote existente
- **Transferir Custodia**: Transfiere la propiedad de un lote a otro address
- **Obtener Información**: Consulta todos los datos públicos de un lote existente
- **Obtener Cadena Blockchain**: Recupera el historial completo de eventos de un contrato
- **Diagnosticar Contrato**: Análisis completo del estado de un contrato
- **Decodificar Input Data**: Utilidades para decodificar transacciones Ethereum

## Endpoints

### GET /api/v1/health
Verifica el estado del microservicio.

**Response:**
```json
{
  "success": true,
  "message": "CrearLoteMicro está funcionando correctamente"
}
```

### GET /api/v1/debug/conexion
Verifica la conexión a la red Sepolia y obtiene información de la blockchain.

**Response:**
```json
{
  "success": true,
  "message": "Conexión a Sepolia exitosa",
  "data": {
    "blockNumber": 4567890,
    "chainId": "11155111"
  }
}
```

### GET /api/v1/lote/info/{contractAddress}
Obtiene toda la información de un lote existente.

**Parámetros de URL:**
- `contractAddress`: Dirección del contrato LoteTracing

**Response:**
```json
{
  "success": true,
  "message": "Información del lote obtenida exitosamente",
  "data": {
    "loteId": "LOTE_MEDICAMENTO_001",
    "fabricante": "0x742d35Cc6634C0532925a3b8D4C9db96590c6C87",
    "propietarioActual": "0x8ba1f109551bD432803012645Hac136c22C177c9",
    "temperaturaMinima": 2,
    "temperaturaMaxima": 8,
    "tempRegMinima": 0,
    "tempRegMaxima": 0,
    "comprometido": false,
    "contractAddress": "0x1234567890123456789012345678901234567890"
  }
}
```

### POST /api/v1/lote/crear
Crea un nuevo lote desplegando un contrato LoteTracing.

**Request Body:**
```json
{
  "loteId": "LOTE001",
  "temperaturaMin": 2,
  "temperaturaMax": 8,
  "walletAddress": "0x...",
  "privateKey": "0x..."
}
```

**Response:**
```json
{
  "success": true,
  "message": "Lote creado exitosamente",
  "data": {
    "contractAddress": "0x...",
    "txHash": "0x...",
    "loteId": "LOTE001"
  },
  "txHash": "0x..."
}
```

### POST /api/v1/lote/temperatura
Registra un rango de temperatura en un lote existente.

**Request Body:**
```json
{
  "contractAddress": "0x...",
  "tempMin": 2,
  "tempMax": 8,
  "walletAddress": "0x...",
  "privateKey": "0x..."
}
```

### POST /api/v1/lote/transferir
Transfiere la custodia de un lote.

**Request Body:**
```json
{
  "contractAddress": "0x...",
  "nuevoPropietario": "0x...",
  "walletAddress": "0x...",
  "privateKey": "0x..."
}
```

### GET /api/v1/lote/cadena/{contractAddress}
Obtiene el historial de eventos (cadena blockchain) de un contrato LoteTracing.

**Parámetros de URL:**
- `contractAddress`: Dirección del contrato LoteTracing

**Limitaciones:**
- Optimizado para RPC gratuitos (busca en los últimos 1000 bloques)
- Para historial completo desde el bloque 0, usar un proveedor RPC de pago

**Response:**
```json
{
  "success": true,
  "message": "Cadena blockchain obtenida exitosamente",
  "data": {
    "contractAddress": "0x1234567890123456789012345678901234567890",
    "loteId": "LOTE_MEDICAMENTO_001",
    "totalEventos": 3,
    "eventos": [
      {
        "tipoEvento": "LoteCreado",
        "blockNumber": 4567890,
        "txHash": "0xabc123...",
        "timestamp": 1640995200,
        "datos": {
          "loteId": "LOTE_MEDICAMENTO_001",
          "fabricante": "0x742d35Cc6634C0532925a3b8D4C9db96590c6C87",
          "temperaturaMinima": 2,
          "temperaturaMaxima": 8
        }
      },
      {
        "tipoEvento": "CustodiaTransferida",
        "blockNumber": 4567920,
        "txHash": "0xdef456...",
        "timestamp": 1640995800,
        "datos": {
          "propietarioAnterior": "0x742d35Cc6634C0532925a3b8D4C9db96590c6C87",
          "nuevoPropietario": "0x8ba1f109551bD432803012645Hac136c22C177c9"
        }
      },
      {
        "tipoEvento": "LoteComprometido",
        "blockNumber": 4567950,
        "txHash": "0x789ghi...",
        "timestamp": 1640996400,
        "datos": {
          "temperaturaRegistrada": 15,
          "motivo": "Temperatura fuera de rango"
        }
      }
    ]
  }
}
```

### GET /api/v1/debug/contrato/{contractAddress}
Realiza un diagnóstico completo del estado de un contrato.

**Parámetros de URL:**
- `contractAddress`: Dirección del contrato a diagnosticar

**Response:**
```json
{
  "success": true,
  "message": "Diagnóstico del contrato completado",
  "data": {
    "contractAddress": "0x...",
    "currentBlock": 4567890,
    "hasCode": true,
    "codeLength": 2048,
    "balance": "0",
    "nonce": 1,
    "contractCalls": {
      "loteId_success": true,
      "loteId_value": "LOTE_001",
      "fabricante_success": true
    },
    "recentActivity": {
      "recentLogs": 3,
      "latestLogBlock": 4567920
    }
  }
}
```

### POST /api/v1/utils/decode
Decodifica el input data de una transacción Ethereum.

**Request Body:**
```json
{
  "inputData": "0xf7b5b4e90000000000000000000000000000000000000000000000000000000000000006"
}
```

**Response:**
```json
{
  "success": true,
  "message": "Input data decodificado exitosamente",
  "data": {
    "functionName": "registrarTemperatura",
    "functionSig": "registrarTemperatura(int8)",
    "parameters": {
      "_temperatura": "6"
    },
    "rawInputData": "0x..."
  }
}
```

### GET /api/v1/utils/decode/specific
Decodifica el input data específico hardcodeado para pruebas.

### GET /api/v1/utils/signatures
Obtiene todas las signatures de funciones del contrato con sus selectores.

**Response:**
```json
{
  "success": true,
  "message": "Signatures de funciones obtenidas exitosamente",
  "data": {
    "f7b5b4e9": "registrarTemperatura(int8)",
    "8da5cb5b": "propietarioActual()",
    "a2fb1175": "loteId()"
  }
}
```

## Configuración

1. Copiar `.env.example` a `.env`
2. Configurar el RPC endpoint de Sepolia
3. Ejecutar el microservicio

## Instalación y Ejecución

```bash
# Instalar dependencias
go mod tidy

# Ejecutar el microservicio
go run main.go
```

## Variables de Entorno

- `SEPOLIA_RPC`: Endpoint RPC de Sepolia
- `PORT`: Puerto del servidor (default: 8080)

## Seguridad

⚠️ **IMPORTANTE**: Las claves privadas se envían en el request body. En producción, considera usar un sistema de gestión de claves más seguro.