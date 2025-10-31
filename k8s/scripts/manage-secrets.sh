#!/bin/bash

# Comprehensive secret management script for Kubernetes
# Usage: ./manage-secrets.sh [command] [options]

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default values
NAMESPACE="default"
COMMAND=""
SECRET_TYPE=""

# Help function
show_help() {
    echo -e "${BLUE}Kubernetes Secret Management Tool${NC}"
    echo ""
    echo "Usage: $0 [command] [options]"
    echo ""
    echo "Commands:"
    echo "  list                    List all secrets in namespace"
    echo "  verify                  Verify secret existence and content"
    echo "  create-blockchain       Create blockchain secrets interactively"
    echo "  delete-blockchain       Delete blockchain secrets"
    echo "  backup                  Backup secrets to file"
    echo "  restore                 Restore secrets from file"
    echo ""
    echo "Options:"
    echo "  -n, --namespace NAME    Kubernetes namespace (default: default)"
    echo "  -h, --help             Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0 list -n medisupply"
    echo "  $0 verify -n default"
    echo "  $0 create-blockchain -n medisupply"
    echo "  $0 backup -n medisupply"
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        list|verify|create-blockchain|delete-blockchain|backup|restore)
            COMMAND="$1"
            shift
            ;;
        -n|--namespace)
            NAMESPACE="$2"
            shift 2
            ;;
        -h|--help)
            show_help
            exit 0
            ;;
        *)
            echo -e "${RED}Unknown option: $1${NC}"
            show_help
            exit 1
            ;;
    esac
done

# Validate command
if [[ -z "$COMMAND" ]]; then
    echo -e "${RED}Error: Command is required${NC}"
    show_help
    exit 1
fi

# Check if kubectl is available
if ! command -v kubectl &> /dev/null; then
    echo -e "${RED}Error: kubectl is not installed or not in PATH${NC}"
    exit 1
fi

# Check if namespace exists
if ! kubectl get namespace "$NAMESPACE" &> /dev/null; then
    echo -e "${YELLOW}Warning: Namespace '$NAMESPACE' does not exist${NC}"
    read -p "Do you want to create it? (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        kubectl create namespace "$NAMESPACE"
        echo -e "${GREEN}✓ Namespace '$NAMESPACE' created${NC}"
    else
        echo -e "${RED}Aborted.${NC}"
        exit 1
    fi
fi

# Command implementations
list_secrets() {
    echo -e "${BLUE}📋 Listing secrets in namespace '$NAMESPACE'${NC}"
    echo ""
    
    if ! kubectl get secrets -n "$NAMESPACE" &> /dev/null; then
        echo -e "${YELLOW}No secrets found in namespace '$NAMESPACE'${NC}"
        return
    fi
    
    kubectl get secrets -n "$NAMESPACE" -o custom-columns="NAME:.metadata.name,TYPE:.type,DATA:.data,AGE:.metadata.creationTimestamp" --sort-by=.metadata.creationTimestamp
    
    echo ""
    echo -e "${BLUE}🔍 Secret details:${NC}"
    for secret in $(kubectl get secrets -n "$NAMESPACE" -o jsonpath='{.items[*].metadata.name}'); do
        echo -e "  ${YELLOW}$secret${NC}:"
        kubectl get secret "$secret" -n "$NAMESPACE" -o jsonpath='{.data}' | jq -r 'keys[]' 2>/dev/null | sed 's/^/    - /' || echo "    (unable to read keys)"
    done
}

verify_secrets() {
    echo -e "${BLUE}🔍 Verifying secrets in namespace '$NAMESPACE'${NC}"
    echo ""
    
    # Check blockchain-secrets
    echo -e "${YELLOW}Checking blockchain-secrets:${NC}"
    if kubectl get secret blockchain-secrets -n "$NAMESPACE" >/dev/null 2>&1; then
        echo "  ✅ blockchain-secrets exists"
        
        # Check required keys
        KEYS=$(kubectl get secret blockchain-secrets -n "$NAMESPACE" -o jsonpath='{.data}' | jq -r 'keys[]' 2>/dev/null || echo "")
        
        for key in "sepolia-rpc" "sepolia-ws" "alchemy-api-key"; do
            if echo "$KEYS" | grep -q "^$key$"; then
                echo "  ✅ $key key found"
            else
                echo "  ❌ $key key missing"
            fi
        done
        
        # Check if values are not empty
        for key in "sepolia-rpc" "sepolia-ws" "alchemy-api-key"; do
            VALUE=$(kubectl get secret blockchain-secrets -n "$NAMESPACE" -o jsonpath="{.data.$key}" 2>/dev/null | base64 -d 2>/dev/null || echo "")
            if [[ -n "$VALUE" ]]; then
                echo "  ✅ $key has value (${#VALUE} chars)"
            else
                echo "  ⚠️  $key is empty or invalid"
            fi
        done
    else
        echo "  ❌ blockchain-secrets not found"
    fi
    
    echo ""
    
    # Check rabbitmq secret
    echo -e "${YELLOW}Checking rabbitmq secret:${NC}"
    if kubectl get secret rabbitmq -n "$NAMESPACE" >/dev/null 2>&1; then
        echo "  ✅ rabbitmq secret exists"
    else
        echo "  ⚠️  rabbitmq secret not found (may not be needed)"
    fi
}

create_blockchain_secrets() {
    echo -e "${BLUE}🔐 Creating blockchain secrets in namespace '$NAMESPACE'${NC}"
    echo ""
    
    # Check if secret already exists
    if kubectl get secret blockchain-secrets -n "$NAMESPACE" >/dev/null 2>&1; then
        echo -e "${YELLOW}Secret blockchain-secrets already exists in namespace '$NAMESPACE'${NC}"
        read -p "Do you want to update it? (y/N): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            echo "Aborted."
            exit 0
        fi
        kubectl delete secret blockchain-secrets -n "$NAMESPACE"
        echo -e "${YELLOW}Existing secret deleted${NC}"
    fi
    
    # Prompt for values
    echo -e "${YELLOW}Please provide the blockchain configuration:${NC}"
    echo ""
    
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
    
    read -p "Alchemy API Key: " ALCHEMY_API_KEY
    if [[ -z "$ALCHEMY_API_KEY" ]]; then
        echo -e "${RED}Error: Alchemy API Key is required${NC}"
        exit 1
    fi
    
    echo ""
    echo -e "${YELLOW}Creating secret with the following configuration:${NC}"
    echo "SEPOLIA_RPC: ${SEPOLIA_RPC:0:50}..."
    echo "SEPOLIA_WS: ${SEPOLIA_WS:0:50}..."
    echo "ALCHEMY_API_KEY: ${ALCHEMY_API_KEY:0:10}..."
    echo ""
    
    # Create the secret
    kubectl create secret generic blockchain-secrets \
        --from-literal=sepolia-rpc="$SEPOLIA_RPC" \
        --from-literal=sepolia-ws="$SEPOLIA_WS" \
        --from-literal=alchemy-api-key="$ALCHEMY_API_KEY" \
        -n "$NAMESPACE"
    
    if [[ $? -eq 0 ]]; then
        echo -e "${GREEN}✓ Secret blockchain-secrets created successfully${NC}"
        echo ""
        echo -e "${BLUE}You can now deploy blockchain services:${NC}"
        echo "# Deploy crear-lote-micro:"
        echo "helm install crear-lote-micro ./k8s/microservice -f k8s/config/services/crear-lote-micro/crear-lote-micro-service-values.yaml -n $NAMESPACE"
        echo "# Deploy alchemy-websocket-micro:"
        echo "helm install alchemy-websocket-micro ./k8s/microservice -f k8s/config/services/alchemy-websocket-micro/alchemy-websocket-micro-service-values.yaml -n $NAMESPACE"
    else
        echo -e "${RED}✗ Failed to create secret${NC}"
        exit 1
    fi
}

delete_blockchain_secrets() {
    echo -e "${YELLOW}⚠️  WARNING: This will delete blockchain secrets in namespace '$NAMESPACE'!${NC}"
    echo ""
    
    if ! kubectl get secret blockchain-secrets -n "$NAMESPACE" >/dev/null 2>&1; then
        echo -e "${YELLOW}Secret blockchain-secrets does not exist in namespace '$NAMESPACE'${NC}"
        exit 0
    fi
    
    read -p "Are you sure you want to delete blockchain-secrets? (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        kubectl delete secret blockchain-secrets -n "$NAMESPACE"
        echo -e "${GREEN}✓ blockchain-secrets deleted${NC}"
    else
        echo -e "${YELLOW}Operation cancelled${NC}"
    fi
}

backup_secrets() {
    echo -e "${BLUE}💾 Backing up secrets from namespace '$NAMESPACE'${NC}"
    
    BACKUP_FILE="secrets-backup-$NAMESPACE-$(date +%Y%m%d-%H%M%S).yaml"
    
    kubectl get secrets -n "$NAMESPACE" -o yaml > "$BACKUP_FILE"
    
    echo -e "${GREEN}✓ Secrets backed up to: $BACKUP_FILE${NC}"
    echo -e "${YELLOW}⚠️  This file contains sensitive data. Store it securely!${NC}"
}

restore_secrets() {
    echo -e "${BLUE}📥 Restoring secrets to namespace '$NAMESPACE'${NC}"
    
    read -p "Enter backup file path: " BACKUP_FILE
    
    if [[ ! -f "$BACKUP_FILE" ]]; then
        echo -e "${RED}Error: Backup file '$BACKUP_FILE' not found${NC}"
        exit 1
    fi
    
    echo -e "${YELLOW}⚠️  This will restore all secrets from the backup file${NC}"
    read -p "Continue? (y/N): " -n 1 -r
    echo
    
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        kubectl apply -f "$BACKUP_FILE" -n "$NAMESPACE"
        echo -e "${GREEN}✓ Secrets restored from: $BACKUP_FILE${NC}"
    else
        echo -e "${YELLOW}Operation cancelled${NC}"
    fi
}

# Execute command
case $COMMAND in
    list)
        list_secrets
        ;;
    verify)
        verify_secrets
        ;;
    create-blockchain)
        create_blockchain_secrets
        ;;
    delete-blockchain)
        delete_blockchain_secrets
        ;;
    backup)
        backup_secrets
        ;;
    restore)
        restore_secrets
        ;;
    *)
        echo -e "${RED}Unknown command: $COMMAND${NC}"
        show_help
        exit 1
        ;;
esac