# Assets del Microservicio CrearLoteMicro

Esta carpeta contiene los archivos de compilación del smart contract y otros recursos necesarios para el funcionamiento del microservicio.

## Estructura

```
assets/
├── contracts/
│   ├── LoteTracing.abi.json      # ABI del contrato LoteTracing
│   ├── LoteTracing.bytecode      # Bytecode del contrato para deployment
│   ├── contract_info.json        # Información y metadatos del contrato
│   ├── loader.go                 # Funciones Go para cargar los assets
│   └── README.md                 # Este archivo
└── README.md
```

## Archivos de Contrato

### LoteTracing.abi.json
Contiene la Application Binary Interface (ABI) del contrato LoteTracing. Este archivo define:
- Funciones disponibles del contrato
- Eventos que emite el contrato
- Tipos de datos de entrada y salida

### LoteTracing.bytecode
Contiene el bytecode compilado del contrato LoteTracing necesario para el deployment en la blockchain.

### contract_info.json
Archivo de metadatos que incluye:
- Información de versión
- Hash del contrato
- Descripción de funcionalidades
- Lista de eventos y funciones

### loader.go
Módulo Go que proporciona funciones para:
- Cargar el ABI del contrato
- Obtener el bytecode
- Validar la integridad de los assets
- Acceder a la información del contrato

## Uso

El microservicio utiliza estos assets a través del módulo `loader.go`:

```go
import "CrearLoteMicro/assets/contracts"

// Obtener ABI
abi := contracts.GetLoteTracingABI()

// Obtener Bytecode
bytecode := contracts.GetLoteTracingBytecode()

// Obtener información del contrato
info, err := contracts.GetContractInfo()

// Validar assets
err := contracts.ValidateContract()
```

## Mantenimiento

### Actualización de Assets

1. **Desde Hardhat**: Copiar desde `smartcontract/lotetracing/artifacts/contracts/LoteTracing.sol/LoteTracing.json`
2. **Actualizar archivos**:
   - Extraer `abi` → `LoteTracing.abi.json`
   - Extraer `bytecode` → `LoteTracing.bytecode`
   - Actualizar `contract_info.json` con nueva información

### Comandos de Actualización

Para automatizar la actualización desde los artifacts de Hardhat:

```bash
# Desde el directorio services/CrearLoteMicro
make update-contract-assets

# O desde el directorio services (raíz)
make update-contract-assets SERVICE=CrearLoteMicro

# Validar assets después de la actualización
make validate-assets

# O desde el directorio services (raíz)
make validate-contract-assets SERVICE=CrearLoteMicro
```

## Ventajas de esta Estructura

1. **Separación de responsabilidades**: Los assets están separados del código de negocio
2. **Mantenibilidad**: Fácil actualización cuando cambia el contrato
3. **Versionado**: Información clara de versiones y cambios
4. **Validación**: Verificación automática de integridad de assets
5. **Embebido**: Los assets se incluyen en el binario compilado (no requiere archivos externos)

## Sincronización con Hardhat

Los assets deben mantenerse sincronizados con la compilación de Hardhat:

- **Fuente**: `smartcontract/lotetracing/artifacts/contracts/LoteTracing.sol/LoteTracing.json`
- **Frecuencia**: Después de cada cambio en el smart contract
- **Validación**: Verificar que el hash coincida con el contrato deployado