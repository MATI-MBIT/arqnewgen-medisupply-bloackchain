# LoteTracing - Sistema de Trazabilidad Farmacéutica

Este proyecto implementa un sistema de trazabilidad para productos farmacéuticos sensibles utilizando Hardhat 3 Beta, con el test runner nativo de Node.js (`node:test`) y la librería `viem` para interacciones con Ethereum.

Para aprender más sobre Hardhat 3 Beta, visita la [Guía de Inicio](https://hardhat.org/docs/getting-started#getting-started-with-hardhat-3). Para compartir feedback, únete a nuestro grupo de [Hardhat 3 Beta](https://hardhat.org/hardhat3-beta-telegram-group) en Telegram o [abre un issue](https://github.com/NomicFoundation/hardhat/issues/new) en nuestro tracker de GitHub.

## Descripción del Proyecto

Este proyecto incluye:

- **LoteTracing Smart Contract**: Sistema completo de trazabilidad para productos farmacéuticos
- **Gestión de Cadena de Frío**: Monitoreo automático de temperatura con sensores IoT
- **Control de Custodia**: Transferencias seguras entre fabricante, distribuidor y farmacia
- **Pruebas Unitarias**: Tests en Solidity compatibles con Foundry
- **Pruebas de Integración**: Tests en TypeScript usando [`node:test`](nodejs.org/api/test.html) y [`viem`](https://viem.sh/)
- **Ejemplos de Despliegue**: Módulos de Ignition para diferentes redes

## Características del Sistema

### Smart Contract LoteTracing

- **Trazabilidad Completa**: Desde fabricación hasta punto de venta
- **Monitoreo de Temperatura**: Registro automático con sensores IoT autorizados
- **Estados del Lote**: Creado, En Tránsito, En Almacén, Comprometido, Entregado
- **Historial Inmutable**: Registro completo de transferencias de custodia
- **Alertas Automáticas**: Marcado automático como comprometido si la temperatura sale del rango

### Actores del Sistema

- **Fabricante**: Crea lotes, autoriza sensores, inicia cadena de custodia
- **Sensores IoT**: Dispositivos autorizados para registrar temperatura
- **Distribuidor**: Intermediario en la cadena de suministro
- **Farmacia**: Punto final de la cadena de distribución

## Uso del Sistema

### Ejecutar Pruebas

Para ejecutar todas las pruebas del proyecto:

```shell
npx hardhat test
```

Ejecutar selectivamente las pruebas de Solidity o `node:test`:

```shell
npx hardhat test solidity
npx hardhat test nodejs
```

### Ejecutar Demo Interactivo

Para ver una demostración completa del sistema de trazabilidad:

```shell
npx hardhat run scripts/demo-lote-tracing.ts
```

### Despliegue del Contrato

#### Despliegue Local

Para desplegar en una cadena local simulada:

```shell
npx hardhat ignition deploy ignition/modules/LoteTracing.ts
```

#### Despliegue en Sepolia

Para desplegar en Sepolia, necesitas una cuenta con fondos. La configuración incluye una Variable de Configuración llamada `SEPOLIA_PRIVATE_KEY`.

Configurar la clave privada usando `hardhat-keystore`:

```shell
npx hardhat keystore set SEPOLIA_PRIVATE_KEY
```

Desplegar en Sepolia:

```shell
npx hardhat ignition deploy --network sepolia ignition/modules/LoteTracing.ts
```

#### Parámetros de Despliegue Personalizados

Puedes personalizar los parámetros del lote durante el despliegue:

```shell
npx hardhat ignition deploy ignition/modules/LoteTracing.ts --parameters '{
  "LoteTracingModule": {
    "sku": "INSULIN-001",
    "loteId": "LOT-2024-001", 
    "temperaturaMinima": 2,
    "temperaturaMaxima": 8,
    "sensorAddress": "0x742d35Cc6634C0532925a3b8D4C9db96c4b4d4d4"
  }
}'
```

## Ejemplos de Uso

### 1. Crear un Nuevo Lote

```typescript
const lote = await viem.deployContract("LoteDeProductoTrazable", [
  "INSULIN-001",           // SKU
  "LOT-2024-001",         // Lote ID
  fechaVencimiento,       // Timestamp de vencimiento
  2,                      // Temperatura mínima (°C)
  8                       // Temperatura máxima (°C)
]);
```

### 2. Autorizar Sensor IoT

```typescript
await lote.write.gestionarSensor([sensorAddress, true], { client: fabricante });
```

### 3. Registrar Temperatura

```typescript
await lote.write.registrarTemperatura([5], { client: sensor });
```

### 4. Transferir Custodia

```typescript
await lote.write.transferirCustodia([distribuidorAddress], { client: fabricante });
```

### 5. Consultar Historial

```typescript
const historial = await lote.read.obtenerHistorialCustodia();
const lecturas = await lote.read.obtenerLecturasTemperatura();
```
