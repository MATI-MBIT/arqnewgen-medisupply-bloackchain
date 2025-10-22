# Prueba del Nuevo Endpoint: Obtener Cadena Blockchain

## Descripción
Este documento describe cómo probar el nuevo endpoint `GET /api/v1/lote/cadena/{contractAddress}` que permite recuperar toda la cadena blockchain de un contrato LoteTracing.

## Limitaciones de RPC Gratuitos
**IMPORTANTE**: Este endpoint está optimizado para trabajar con proveedores RPC gratuitos (como Infura o Alchemy) que tienen limitaciones:
- Máximo 10 bloques por consulta `eth_getLogs`
- Búsqueda limitada a los últimos 1000 bloques para evitar timeouts
- Si necesitas historial completo desde el bloque 0, considera usar un plan de pago del proveedor RPC

## Flujo de Prueba Recomendado

### 1. Crear un Lote
```bash
POST /api/v1/lote/crear
{
    "loteId": "LOTE_TEST_001",
    "temperaturaMin": 2,
    "temperaturaMax": 8,
    "walletAddress": "0x742d35Cc6634C0532925a3b8D4C9db96590c6C87",
    "privateKey": "0x4c0883a69102937d6231471b5dbb6204fe5129617082792ae468d01a3f362318"
}
```

### 2. Registrar Temperatura Normal
```bash
POST /api/v1/lote/temperatura
{
    "contractAddress": "{contract_address_from_step_1}",
    "temperatura": 5,
    "walletAddress": "0x742d35Cc6634C0532925a3b8D4C9db96590c6C87",
    "privateKey": "0x4c0883a69102937d6231471b5dbb6204fe5129617082792ae468d01a3f362318"
}
```

### 3. Transferir Custodia
```bash
POST /api/v1/lote/transferir
{
    "contractAddress": "{contract_address_from_step_1}",
    "nuevoPropietario": "0x8ba1f109551bD432803012645Hac136c22C177c9",
    "walletAddress": "0x742d35Cc6634C0532925a3b8D4C9db96590c6C87",
    "privateKey": "0x4c0883a69102937d6231471b5dbb6204fe5129617082792ae468d01a3f362318"
}
```

### 4. Registrar Temperatura Fuera de Rango (Opcional)
```bash
POST /api/v1/lote/temperatura
{
    "contractAddress": "{contract_address_from_step_1}",
    "temperatura": 15,
    "walletAddress": "0x8ba1f109551bD432803012645Hac136c22C177c9",
    "privateKey": "{private_key_of_new_owner}"
}
```

### 5. Obtener Cadena Blockchain Completa
```bash
GET /api/v1/lote/cadena/{contract_address_from_step_1}
```

## Respuesta Esperada

La respuesta del paso 5 debería incluir todos los eventos generados en los pasos anteriores:

```json
{
    "success": true,
    "message": "Cadena blockchain obtenida exitosamente",
    "data": {
        "contractAddress": "0x...",
        "loteId": "LOTE_TEST_001",
        "totalEventos": 3,
        "eventos": [
            {
                "tipoEvento": "LoteCreado",
                "blockNumber": 123456,
                "txHash": "0x...",
                "timestamp": 1640995200,
                "datos": {
                    "loteId": "LOTE_TEST_001",
                    "fabricante": "0x742d35Cc6634C0532925a3b8D4C9db96590c6C87",
                    "temperaturaMinima": 2,
                    "temperaturaMaxima": 8
                }
            },
            {
                "tipoEvento": "CustodiaTransferida",
                "blockNumber": 123457,
                "txHash": "0x...",
                "timestamp": 1640995800,
                "datos": {
                    "propietarioAnterior": "0x742d35Cc6634C0532925a3b8D4C9db96590c6C87",
                    "nuevoPropietario": "0x8ba1f109551bD432803012645Hac136c22C177c9"
                }
            }
        ]
    }
}
```

## Notas Importantes

1. **Orden de Eventos**: Los eventos aparecen en orden cronológico según el número de bloque
2. **Timestamps**: Los timestamps están en formato Unix (segundos desde epoch)
3. **Datos Específicos**: Cada tipo de evento tiene datos específicos en el campo `datos`
4. **Historial Inmutable**: Una vez registrados en la blockchain, los eventos no pueden modificarse
5. **Limitaciones de Búsqueda**: Por limitaciones de RPC gratuitos, solo busca en los últimos 1000 bloques
6. **Optimización**: El endpoint verifica primero que el contrato exista antes de buscar eventos

## Solución a Errores de RPC

Si encuentras errores como:
```
"Under the Free tier plan, you can make eth_getLogs requests with up to a 10 block range"
```

El endpoint ahora usa una estrategia optimizada que:
- Busca solo en los últimos 1000 bloques
- Usa lotes de máximo 10 bloques
- Maneja errores de manera elegante
- Continúa la búsqueda aunque algunos lotes fallen

## Casos de Uso

- **Auditoría**: Verificar todo el historial de un lote
- **Trazabilidad**: Seguir la cadena de custodia completa
- **Compliance**: Demostrar el cumplimiento de la cadena de frío
- **Investigación**: Analizar incidentes o problemas de calidad