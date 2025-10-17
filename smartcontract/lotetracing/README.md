# LoteTracing PoC - Prueba de Concepto de Trazabilidad

Este proyecto implementa una **Prueba de Concepto (PoC)** simplificada para trazabilidad de productos farmacéuticos utilizando Hardhat 3 Beta, con el test runner nativo de Node.js (`node:test`) y la librería `viem` para interacciones con Ethereum.

Para aprender más sobre Hardhat 3 Beta, visita la [Guía de Inicio](https://hardhat.org/docs/getting-started#getting-started-with-hardhat-3). Para compartir feedback, únete a nuestro grupo de [Hardhat 3 Beta](https://hardhat.org/hardhat3-beta-telegram-group) en Telegram o [abre un issue](https://github.com/NomicFoundation/hardhat/issues/new) en nuestro tracker de GitHub.

## Descripción del Proyecto

Esta PoC se centra en los aspectos fundamentales de la trazabilidad:

- **LoteTracing PoC Smart Contract**: Implementación simplificada de trazabilidad
- **Gestión de Cadena de Frío**: Monitoreo básico de temperatura por propietario
- **Control de Custodia**: Transferencias simples entre actores de la cadena
- **Pruebas Unitarias**: Tests en Solidity compatibles con Foundry
- **Pruebas de Integración**: Tests en TypeScript usando [`node:test`](nodejs.org/api/test.html) y [`viem`](https://viem.sh/)
- **Ejemplos de Despliegue**: Módulos de Ignition para diferentes redes

## Características de la PoC

### Smart Contract LoteTracing PoC

- **Trazabilidad Básica**: Seguimiento de propietario actual y estado de integridad
- **Monitoreo de Temperatura**: Registro por propietario actual únicamente
- **Estado Binario**: Íntegro o Comprometido (simplificado)
- **Eventos Inmutables**: Registro de creación, transferencias y compromisos
- **Control de Acceso**: Solo el propietario actual puede registrar temperaturas

### Actores del Sistema

- **Fabricante**: Crea el lote e inicia la cadena de custodia
- **Distribuidor**: Intermediario en la cadena de suministro
- **Farmacia**: Punto final de la cadena de distribución

### Simplificaciones de la PoC

- No hay sensores IoT separados (el propietario registra temperaturas)
- Estados simplificados (solo íntegro/comprometido)
- Sin historial detallado de lecturas (solo eventos)
- Sin fechas de vencimiento o SKUs complejos

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

Desplegar en Sepolia usando Ignition:

```shell
npx hardhat ignition deploy --network sepolia ignition/modules/LoteTracing.ts
```

O usar el script de despliegue simplificado:

```shell
npx hardhat run scripts/deploy-sepolia.ts --network sepolia
```

#### Parámetros de Despliegue Personalizados

Puedes personalizar los parámetros del lote durante el despliegue:

```shell
npx hardhat ignition deploy ignition/modules/LoteTracing.ts --parameters '{
  "LoteTracingModule": {
    "loteId": "LOT-2024-001",
    "temperaturaMinima": 2,
    "temperaturaMaxima": 8
  }
}'
```

## Ejemplos de Uso

### 1. Crear un Nuevo Lote

```typescript
const lote = await viem.deployContract("LoteDeProductoTrazablePoC", [
  "LOT-2024-001", // Lote ID
  2, // Temperatura mínima (°C)
  8, // Temperatura máxima (°C)
]);
```

### 2. Registrar Temperatura

```typescript
await lote.write.registrarTemperatura([5]); // Solo el propietario actual
```

### 3. Transferir Custodia

```typescript
await lote.write.transferirCustodia([nuevoPropietarioAddress]);
```

### 4. Consultar Estado

```typescript
const comprometido = await lote.read.comprometido();
const propietario = await lote.read.propietarioActual();
const fabricante = await lote.read.fabricante();
```

## Arquitectura Simplificada

```text
┌─────────────┐    registrarTemperatura()   ┌─────────────────┐
│ Fabricante  │ ──────────────────────────► │ Smart Contract  │
└─────────────┘                             │ LoteTracing PoC │
                                            └─────────────────┘
┌─────────────┐    transferirCustodia()             │
│ Distribuidor│ ◄───────────────────────────────────┘
└─────────────┘                                     │
                                                    │
┌─────────────┐    registrarTemperatura()           │
│ Farmacia    │ ◄───────────────────────────────────┘
└─────────────┘
```

## Eventos del Contrato

- `LoteCreado`: Emitido al crear un nuevo lote
- `CustodiaTransferida`: Emitido al transferir la custodia
- `LoteComprometido`: Emitido cuando la temperatura sale del rango permitido

## Limitaciones de la PoC

Esta es una implementación simplificada para demostrar conceptos básicos:

- **Sin persistencia de lecturas**: Solo se almacena el estado comprometido
- **Sin sensores IoT**: El propietario registra manualmente las temperaturas
- **Sin historial detallado**: Solo eventos para trazabilidad básica
- **Estados binarios**: Solo íntegro o comprometido
- **Sin validaciones complejas**: Implementación mínima para PoC

## Próximos Pasos

Para una implementación completa se podría considerar:

- Integración con sensores IoT reales
- Historial detallado de lecturas de temperatura
- Estados más granulares del lote
- Integración con sistemas de gestión de inventario
- Interfaz web para visualización de datos
- Notificaciones automáticas por compromisos
