#!/bin/bash

# Script para ejecutar docker-compose con variables de entorno
# Uso: ./run.sh [comando docker-compose]

if [ -z "$SEPOLIA_RPC" ]; then
    echo "Error: La variable SEPOLIA_RPC no está configurada"
    echo "Ejecuta: export SEPOLIA_RPC='tu_url_aqui'"
    exit 1
fi

echo "Usando SEPOLIA_RPC: ${SEPOLIA_RPC:0:30}..."

# Ejecutar docker-compose con la variable exportada
SEPOLIA_RPC="$SEPOLIA_RPC" docker-compose "$@"