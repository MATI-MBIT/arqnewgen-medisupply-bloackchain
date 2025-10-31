#!/bin/bash

# Generic service deployment script with Istio injection
# Usage: ./deploy-service.sh <service-name> [namespace] [release-name] [image-tag]

set -e  # Exit on any error

SERVICE_NAME=${1}
NAMESPACE=${2:-default}
RELEASE_NAME=${3:-$SERVICE_NAME}
IMAGE_TAG=${4:-latest}
CHART_PATH="microservice"
VALUES_FILE="config/services/$SERVICE_NAME/$SERVICE_NAME-service-values.yaml"

# Validate inputs
if [ -z "$SERVICE_NAME" ]; then
    echo "❌ Error: Service name is required"
    echo "Usage: $0 <service-name> [namespace] [release-name] [image-tag]"
    echo "Example: $0 auth default auth-service v1.0.0"
    exit 1
fi

if [ ! -f "$VALUES_FILE" ]; then
    echo "❌ Error: Values file not found: $VALUES_FILE"
    exit 1
fi

if [ ! -d "$CHART_PATH" ]; then
    echo "❌ Error: Chart path not found: $CHART_PATH"
    exit 1
fi

echo "🚀 Deploying $SERVICE_NAME Service..."
echo "📦 Namespace: $NAMESPACE"
echo "🏷️  Release: $RELEASE_NAME"
echo "🐳 Image Tag: $IMAGE_TAG"
echo "📋 Values: $VALUES_FILE"
echo "📊 Chart: $CHART_PATH"
echo ""

# Check dependencies
for cmd in kubectl helm; do
    if ! command -v $cmd &> /dev/null; then
        echo "❌ $cmd is not installed or not in PATH"
        exit 1
    fi
done

# Setup namespace with Istio injection
echo "🔧 Setting up namespace $NAMESPACE..."
kubectl create namespace $NAMESPACE --dry-run=client -o yaml | kubectl apply -f -

echo "🕸️  Enabling Istio injection for namespace $NAMESPACE..."
kubectl label namespace $NAMESPACE istio-injection=enabled --overwrite

# Verify Istio injection
ISTIO_LABEL=$(kubectl get namespace $NAMESPACE -o jsonpath='{.metadata.labels.istio-injection}' 2>/dev/null || echo "")
if [ "$ISTIO_LABEL" = "enabled" ]; then
    echo "✅ Istio injection enabled for namespace $NAMESPACE"
else
    echo "⚠️  Warning: Istio injection may not be properly configured"
fi

# Check for required secrets before deployment
echo "🔐 Checking for required secrets..."
if [ "$SERVICE_NAME" = "crear-lote-micro" ]; then
    if ! kubectl get secret blockchain-secrets -n $NAMESPACE >/dev/null 2>&1; then
        echo "❌ Error: Required secret 'blockchain-secrets' not found in namespace $NAMESPACE"
        echo "Please create the secret first using:"
        echo "  make create-secrets-interactive"
        echo "  OR"
        echo "  ./scripts/create-blockchain-secrets.sh $NAMESPACE"
        exit 1
    else
        echo "✅ Required secret 'blockchain-secrets' found"
    fi
fi

# Deploy using Helm
echo ""
echo "📦 Deploying with Helm..."
helm upgrade --install $RELEASE_NAME $CHART_PATH \
  --namespace $NAMESPACE \
  --values $VALUES_FILE \
  --set image.tag=$IMAGE_TAG \
  --wait \
  --timeout 300s

if [ $? -eq 0 ]; then
  echo ""
  echo "✅ $SERVICE_NAME Service deployed successfully!"
  echo ""
  echo "📋 Deployment info:"
  kubectl get pods,svc,ingress -n $NAMESPACE -l app.kubernetes.io/instance=$RELEASE_NAME
  echo ""
  echo "🔍 Useful commands:"
  echo "  Logs: kubectl logs -n $NAMESPACE -l app.kubernetes.io/instance=$RELEASE_NAME -f"
  echo "  Port-forward: kubectl port-forward -n $NAMESPACE svc/$RELEASE_NAME 8080:8080"
  echo "  Delete: helm uninstall $RELEASE_NAME -n $NAMESPACE"
  echo ""
  
  # Show ingress info if available
  INGRESS_HOST=$(kubectl get ingress -n $NAMESPACE -l app.kubernetes.io/instance=$RELEASE_NAME -o jsonpath='{.items[0].spec.rules[0].host}' 2>/dev/null || echo "")
  if [ -n "$INGRESS_HOST" ]; then
    echo "🌐 Service endpoints:"
    echo "  External: http://$INGRESS_HOST"
    echo "  Health: http://$INGRESS_HOST/health"
  fi
else
  echo "❌ Deployment failed!"
  exit 1
fi