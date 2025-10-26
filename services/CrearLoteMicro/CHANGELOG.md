# Changelog - CrearLoteMicro

## Versión 2.1.0 - Actualización Final del Contrato LoteTracing

### Cambios Principales

#### Contrato Inteligente Actualizado

- **Nuevos campos de estado**: `tempRegMinima` y `tempRegMaxima` para rastrear las temperaturas registradas
- **Función `registrarTemperatura` simplificada**: Ahora acepta solo 2 parámetros (`_tempMin`, `_tempMax`)
- **Sin restricción de propietario**: La función `registrarTemperatura` puede ser llamada por cualquier dirección
- **Evento `LoteComprometido` mejorado**: Incluye `propietario`, `tempMin`, `tempMax`, `comprometido` y `motivo`
- **Evento `CustodiaTransferida` actualizado**: Incluye el estado `comprometido` del lote

#### API del Microservicio

- **Endpoint `POST /api/v1/lote/temperatura`**:
  - Ahora requiere solo campos: `tempMin` y `tempMax` (eliminado `temperatura`)
  - Registra rangos de temperatura en lugar de valores específicos
  - No requiere ser el propietario del lote para registrar temperaturas
- **Endpoint `GET /api/v1/lote/info/{contractAddress}`**:
  - Respuesta incluye nuevos campos: `tempRegMinima` y `tempRegMaxima`
  - Muestra las últimas temperaturas registradas

#### Modelos de Datos

- **`RegistrarTemperaturaRequest`**: Eliminado campo `temperatura`, mantiene solo `tempMin` y `tempMax`
- **`LoteInfoResponse`**: Agregados campos `tempRegMinima` y `tempRegMaxima`

#### Procesamiento de Eventos

- **Evento `CustodiaTransferida`**: Ahora decodifica el campo `comprometido`
- **Evento `LoteComprometido`**: Decodifica `propietario`, `tempMin`, `tempMax`, `comprometido` y `motivo`

#### Utilidades de Decodificación

- **ABI actualizado**: Incluye las nuevas funciones y eventos
- **Ejemplo de input data**: Actualizado para mostrar la función con 2 parámetros
- **Diagnóstico de contratos**: Incluye verificación de los nuevos campos

#### Colección de Postman

- **Requests actualizados**: Eliminado campo `temperatura`, mantiene solo `tempMin` y `tempMax`
- **Ejemplos de respuesta**: Actualizados con los nuevos campos
- **Documentación**: Mejorada para reflejar los cambios en la API
- **Ejemplo fuera de rango**: Actualizado para mostrar rangos en lugar de temperatura específica

### Compatibilidad

- **Breaking Changes**: Esta versión NO es compatible con contratos desplegados con la versión anterior
- **Nuevos contratos**: Deben ser desplegados con la nueva versión del contrato
- **API**: Los endpoints existentes requieren parámetros adicionales

### Migración

Para migrar a esta versión:

1. Desplegar nuevos contratos usando la versión actualizada
2. Actualizar las aplicaciones cliente para eliminar el campo `temperatura` y usar solo `tempMin` y `tempMax`
3. Actualizar la lógica de negocio para trabajar con rangos de temperatura en lugar de valores específicos
4. Usar la colección de Postman actualizada para pruebas

### Archivos Modificados

- `services/blockchain_service.go`: ABI y bytecode actualizados, función `RegistrarTemperatura` modificada
- `models/lote_models.go`: Modelos actualizados con nuevos campos
- `handlers/lote_handlers.go`: Handler actualizado para nuevos parámetros
- `utils/decoder.go`: ABI y ejemplos actualizados
- `README.md`: Documentación actualizada
- `CrearLoteMicro.postman_collection.json`: Colección actualizada
- `CrearLoteMicro.postman_environment.json`: Variables de entorno actualizadas

## Vers

ión 2.1.1 - Actualización de Bytecode Final

### Fecha: 2024-12-XX

#### Cambios Realizados

- **Bytecode Actualizado**: Se actualizó el bytecode del contrato en `blockchain_service.go` para reflejar la versión más reciente del contrato
- **Lógica de Compromiso Corregida**: El bytecode ahora incluye la lógica corregida que usa `<=` y `>=` en lugar de `<` y `>`
- **Hash del Contrato**: `0x455c69ef9b29f1f480ef86ce2ec4998483cf555a7da36512be6981ab5c4620f4`

#### Archivos Modificados

- `services/blockchain_service.go`: Constante `contractBytecode` actualizada
- `BYTECODE_ACTUALIZADO.md`: Documentación de la actualización

#### Verificación

- ✅ Bytecode sincronizado con el contrato compilado más reciente
- ✅ Lógica de compromiso corregida implementada
- ✅ Microservicio listo para deploys con la nueva lógica

#### Impacto

- Los nuevos deploys del contrato usarán la lógica de compromiso corregida
- Las pruebas de temperatura se comportarán según los operadores `<=` y `>=`
- Compatibilidad completa con la versión más reciente del contrato Solidity

## Versión 2.1.2 - CORRECCIÓN CRÍTICA de Lógica de Compromiso

### Fecha: 2024-12-XX

#### Cambios Críticos Realizados

- **CORRECCIÓN DE LÓGICA CRÍTICA**: Se corrigió la lógica de compromiso del contrato que estaba invertida
- **Bytecode Actualizado**: Se actualizó el bytecode del contrato con la lógica corregida
- **Lógica Anterior (INCORRECTA)**: `if (temperaturaMinima <= _tempMin || temperaturaMaxima >= _tempMax)`
- **Lógica Nueva (CORRECTA)**: `if (_tempMin < temperaturaMinima || _tempMax > temperaturaMaxima)`

#### Impacto del Cambio

- **Antes**: El lote se comprometía cuando las temperaturas estaban DENTRO del rango permitido (comportamiento incorrecto)
- **Ahora**: El lote se compromete solo cuando las temperaturas están FUERA del rango permitido (comportamiento correcto)

#### Archivos Modificados

- `services/blockchain_service.go`: Bytecode actualizado con lógica corregida
- `BYTECODE_ACTUALIZADO.md`: Documentación actualizada con la corrección

#### Verificación

- ✅ Lógica de compromiso corregida e implementada
- ✅ Bytecode sincronizado con el contrato compilado más reciente
- ✅ Comportamiento correcto para sistema de trazabilidad de cadena de frío
- ✅ Hash del nuevo contrato: `0xbe04fcc7a8591026bdc9f74276ee287097af5e2884636657ed05104a5501ad7a`

#### Ejemplo de Corrección

**Caso**: tempMin=4, tempMax=6, rangoPermitido=2-8

- **Antes**: Se comprometía incorrectamente (4-6 dentro de 2-8)
- **Ahora**: NO se compromete correctamente (4-6 dentro de 2-8)
