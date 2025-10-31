#!/bin/bash

# Script to create blockchain secrets for crear-lote-micro service
# Usage: ./create-blockchain-secrets.sh [namespace]

set -e

NAMESPACE=${1:-default}
SECRET_NAME="blockchain-secrets"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}Creating blockchain secrets for crear-lote-micro service${NC}"
echo "Namespace: $NAMESPACE"
echo "Secret name: $SECRET_NAME"
echo

# Check if secret already exists
if kubectl get secret $SECRET_NAME -n $NAMESPACE >/dev/null 2>&1; then
    echo -e "${YELLOW}Secret $SECRET_NAME already exists in namespace $NAMESPACE${NC}"
    read -p "Do you want to update it? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "Aborted."
        exit 0
    fi
    DELETE_EXISTING=true
else
    DELETE_EXISTING=false
fi

# Prompt for values
echo -e "${YELLOW}Please provide the blockchain endpoints:${NC}"
echo

read -p "Sepolia RPC endpoint (e.g., https://eth-sepolia.g.alchemy.com/v2/YOUR_KEY): " SEPOLIA_RPC
if [[ -z "$SEPOLIA_RPC" ]]; then
    echo -e "${RED}Error: Sepolia RPC endpoint is required${NC}"
    exit 1
fi

read -p "Sepolia WebSocket endpoint (e.g., wss://eth-sepolia.g.alchemy.com/v2/YOUR_KEY): " SEPOLIA_WS
if [[ -z "$SEPOLIA_WS" ]]; then
    echo -e "${RED}Error: Sepolia WebSocket endpoint is required${NC}"
    exit 1
fi

echo
echo -e "${YELLOW}Creating secret with the following values:${NC}"
echo "SEPOLIA_RPC: ${SEPOLIA_RPC:0:50}..."
echo "SEPOLIA_WS: ${SEPOLIA_WS:0:50}..."
echo

# Delete existing secret if needed
if [[ "$DELETE_EXISTING" == "true" ]]; then
    echo -e "${YELLOW}Deleting existing secret...${NC}"
    kubectl delete secret $SECRET_NAME -n $NAMESPACE
fi

# Create the secret
echo -e "${YELLOW}Creating secret...${NC}"
kubectl create secret generic $SECRET_NAME \
    --from-literal=sepolia-rpc="$SEPOLIA_RPC" \
    --from-literal=sepolia-ws="$SEPOLIA_WS" \
    -n $NAMESPACE

if [[ $? -eq 0 ]]; then
    echo -e "${GREEN}✓ Secret $SECRET_NAME created successfully in namespace $NAMESPACE${NC}"
    echo
    echo -e "${YELLOW}You can now deploy the crear-lote-micro service:${NC}"
    echo "helm install crear-lote-micro ./k8s/microservice -f k8s/config/services/crear-lote-micro/crear-lote-micro-service-values.yaml -n $NAMESPACE"
else
    echo -e "${RED}✗ Failed to create secret${NC}"
    exit 1
fi

echo
echo -e "${YELLOW}To verify the secret was created:${NC}"
echo "kubectl get secret $SECRET_NAME -n $NAMESPACE -o yaml"
echo
echo -e "${YELLOW}To view secret keys (without values):${NC}"
echo "kubectl get secret $SECRET_NAME -n $NAMESPACE -o jsonpath='{.data}' | jq 'keys'"