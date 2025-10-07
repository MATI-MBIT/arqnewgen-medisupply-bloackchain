# KEDA Scaling Makefile - MediSupply EDA

.PHONY: help install uninstall status clean

# Variables
MEDISUPPLY_NS=medisupply
MEDIORDER_NS=mediorder
MEDILOGISTIC_NS=medilogistic
KEDA_NS=keda-system

help: ## Show this help message
	@echo "KEDA Scaling Management - MediSupply EDA"
	@echo ""
	@echo "Available commands:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

check-keda: ## Check if KEDA is installed
	@echo "Checking KEDA installation..."
	@kubectl get pods -n $(KEDA_NS) 2>/dev/null || (echo "❌ KEDA not found. Install KEDA first: make -C ../keda install" && exit 1)
	@echo "✅ KEDA is running"

check-namespaces: ## Check if required namespaces exist
	@echo "Checking namespaces..."
	@kubectl get namespace $(MEDISUPPLY_NS) >/dev/null 2>&1 || kubectl create namespace $(MEDISUPPLY_NS)
	@kubectl get namespace $(MEDIORDER_NS) >/dev/null 2>&1 || kubectl create namespace $(MEDIORDER_NS)
	@kubectl get namespace $(MEDILOGISTIC_NS) >/dev/null 2>&1 || kubectl create namespace $(MEDILOGISTIC_NS)
	@echo "✅ Namespaces ready"

install-scalers: check-keda check-namespaces ## Install all ScaledObjects
	@echo "Installing KEDA ScaledObjects..."
	@kubectl apply -f mqtt-order-event-client-scaler.yaml
	@kubectl apply -f order-service-scaler.yaml
	@kubectl apply -f warehouse-batch-scaler.yaml
	@echo "✅ All ScaledObjects installed"

install: install-scalers ## Install all KEDA scaling configurations
	@echo "🚀 KEDA scaling configurations installed successfully!"

uninstall: ## Remove all KEDA scaling configurations
	@echo "Removing KEDA ScaledObjects..."
	@kubectl delete -f mqtt-order-event-client-scaler.yaml --ignore-not-found=true
	@kubectl delete -f order-service-scaler.yaml --ignore-not-found=true
	@kubectl delete -f warehouse-batch-scaler.yaml --ignore-not-found=true
	@echo "✅ KEDA scaling configurations removed"

status: ## Show status of all ScaledObjects and HPAs
	@echo "=== KEDA ScaledObjects ==="
	@kubectl get scaledobjects -A -o wide 2>/dev/null || echo "No ScaledObjects found"
	@echo ""
	@echo "=== Generated HPAs ==="
	@kubectl get hpa -A -o wide 2>/dev/null || echo "No HPAs found"
	@echo ""


logs: ## Show KEDA operator logs
	@echo "=== KEDA Operator Logs ==="
	@kubectl logs deployment/keda-operator -n $(KEDA_NS) --tail=50

metrics: ## Show available external metrics
	@echo "=== Available External Metrics ==="
	@kubectl get --raw "/apis/external.metrics.k8s.io/v1beta1" | jq . 2>/dev/null || echo "External metrics API not available"

describe: ## Describe all ScaledObjects (for debugging)
	@echo "=== ScaledObject Details ==="
	@kubectl describe scaledobject mqtt-order-event-client-scaler -n $(MEDILOGISTIC_NS) 2>/dev/null || echo "mqtt-order-event-client-scaler not found"
	@kubectl describe scaledobject order-service-scaler -n $(MEDILOGISTIC_NS) 2>/dev/null || echo "order-service-scaler not found"
	@kubectl describe scaledobject warehouse-batch-scaler -n $(MEDILOGISTIC_NS) 2>/dev/null || echo "warehouse-batch-scaler not found"
	@echo ""
	@echo "=== Target Deployments ==="
	@kubectl describe deployment mqtt-order-event-client-microservice -n $(MEDILOGISTIC_NS) 2>/dev/null || echo "mqtt-order-event-client-microservice deployment not found"
	@kubectl describe deployment order-service-microservice -n $(MEDILOGISTIC_NS) 2>/dev/null || echo "order-service-microservice deployment not found"
	@kubectl describe deployment warehouse-batch-microservice -n $(MEDILOGISTIC_NS) 2>/dev/null || echo "warehouse-batch-microservice deployment not found"

clean: uninstall ## Remove everything
	@echo "🧹 All KEDA scaling resources cleaned up"

test-kafka: ## Test Kafka connectivity (requires kafkacat/kcat)
	@echo "Testing Kafka connectivity..."
	@kubectl run kafka-test --rm -i --tty --image=confluentinc/cp-kafkacat:latest --restart=Never -- \
		kafkacat -b kafka.medisupply.svc.cluster.local:9092 -L

validate: ## Validate all YAML files
	@echo "Validating YAML files..."
	@for file in *.yaml; do \
		echo "Validating $$file..."; \
		kubectl apply --dry-run=client -f $$file >/dev/null && echo "✅ $$file is valid" || echo "❌ $$file has errors"; \
	done