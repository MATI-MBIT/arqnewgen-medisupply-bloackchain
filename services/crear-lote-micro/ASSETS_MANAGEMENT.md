# Gestión de Assets del Smart Contract

Este documento explica cómo gestionar los assets del smart contract LoteTracing en el microservicio CrearLoteMicro.

## Estructura de Assets

Los assets del contrato se encuentran organizados en:

```
assets/
├── contracts/
│   ├── LoteTracing.abi.json      # ABI del contrato
│   ├── LoteTracing.bytecode      # Bytecode para deployment
│   ├── contract_info.json        # Metadatos del contrato
│   └── loader.go                 # Módulo Go para cargar assets
└── README.md
```

## Comandos Disponibles

### Desde el directorio CrearLoteMicro

```bash
cd services/CrearLoteMicro

# Mostrar ayuda
make help

# Actualizar assets desde Hardhat
make update-contract-assets

# Validar integridad de assets
make validate-assets

# Mostrar información del contrato
make show-contract-info

# Configuración completa para desarrollo
make dev-setup

# Actualizar y recompilar
make dev-update
```

### Desde el directorio services (raíz)

```bash
cd services

# Actualizar assets del contrato
make update-contract-assets SERVICE=CrearLoteMicro

# Validar assets del contrato
make validate-contract-assets SERVICE=CrearLoteMicro
```

## Flujo de Trabajo

### 1. Después de cambios en el Smart Contract

Cuando se modifica el contrato LoteTracing en Hardhat:

```bash
# 1. Compilar el contrato en Hardhat
cd smartcontract/lotetracing
npx hardhat compile

# 2. Actualizar assets en el microservicio
cd ../../services/CrearLoteMicro
make update-contract-assets

# 3. Verificar que todo compile correctamente
make build
```

### 2. Validación de Assets

Para verificar que los assets están correctos:

```bash
# Validar integridad
make validate-assets

# Mostrar información actual
make show-contract-info
```

### 3. Desarrollo Local

Para configurar el entorno de desarrollo:

```bash
# Configuración completa (actualiza assets + compila)
make dev-setup

# Solo actualizar y recompilar
make dev-update
```

## Automatización

### Integración con CI/CD

En pipelines de CI/CD, puedes usar:

```yaml
# Ejemplo para GitHub Actions
- name: Update Contract Assets
  run: |
    cd services
    make update-contract-assets SERVICE=CrearLoteMicro
    make validate-contract-assets SERVICE=CrearLoteMicro
```

### Pre-commit Hooks

Para automatizar la sincronización:

```bash
# .git/hooks/pre-commit
#!/bin/bash
cd services/CrearLoteMicro
make validate-assets || {
    echo "❌ Assets del contrato no están sincronizados"
    echo "💡 Ejecuta: make update-contract-assets"
    exit 1
}
```

## Troubleshooting

### Error: Artifact de Hardhat no encontrado

```bash
❌ Error: No se encontró el artifact de Hardhat
```

**Solución**: Compilar el contrato en Hardhat:
```bash
cd smartcontract/lotetracing
npx hardhat compile
```

### Error: jq no está instalado

```bash
❌ jq no está instalado
```

**Solución**: Instalar jq:
```bash
# macOS
brew install jq

# Ubuntu/Debian
sudo apt-get install jq

# CentOS/RHEL
sudo yum install jq
```

### Error: El módulo Go no compila

```bash
❌ Error: El módulo de assets no compila
```

**Solución**: Verificar que los assets tengan formato correcto:
```bash
make validate-assets
```

## Verificación Manual

Para verificar manualmente que los assets están sincronizados:

```bash
# 1. Comparar hash del bytecode
cd services/CrearLoteMicro
HARDHAT_HASH=$(jq -r '.bytecode' ../../smartcontract/lotetracing/artifacts/contracts/LoteTracing.sol/LoteTracing.json | tail -c 65 | head -c 64)
ASSETS_HASH=$(jq -r '.hash' assets/contracts/contract_info.json | tail -c 64)

if [ "$HARDHAT_HASH" = "$ASSETS_HASH" ]; then
    echo "✅ Assets sincronizados"
else
    echo "❌ Assets desincronizados"
fi

# 2. Verificar compilación
go build -o /tmp/test ./assets/contracts/loader.go && echo "✅ Assets válidos" || echo "❌ Assets inválidos"
```

## Mejores Prácticas

1. **Siempre actualizar después de cambios**: Ejecutar `make update-contract-assets` después de modificar el smart contract
2. **Validar antes de commit**: Usar `make validate-assets` antes de hacer commit
3. **Documentar cambios**: Actualizar `contract_info.json` con información relevante
4. **Probar compilación**: Verificar que `make build` funcione después de actualizar assets
5. **Sincronizar en equipo**: Asegurar que todos los desarrolladores tengan los mismos assets