# Makefile para gestión de cluster local con Kind/Minikube y charts de Helm
# Requiere: kind/minikube, kubectl, helm, docker

CLUSTER_NAME := local-k8s
KUBECONFIG_PATH := ~/.kube/config
PROVIDER := $(or $(PROVIDER),kind)

.PHONY: help init deploy clean destroy kiali kafka-ui mqtt rabbitmq kafka-rabbitmq-replicator kafka-rabbitmq-replicator-logs kafka-rabbitmq-replicator-status create-kafka-topics wait-for-kafka create-topic-if-not-exists create-rabbitmq-resources wait-for-rabbitmq

help: ## Mostrar ayuda
	@echo "Comandos disponibles:"
	@echo "  init PROVIDER=kind     - Crear cluster local con Kind (default)"
	@echo "  init PROVIDER=minikube - Crear cluster local con Minikube"
	@echo "  deploy                 - Desplegar toda la arquitectura EDA completa"
	@echo "  status                 - Mostrar estado de todos los componentes"
	@echo "  kiali                  - Abrir Kiali dashboard (requiere Istio)"
	@echo ""
	@echo "  clean                  - Eliminar charts del cluster"
	@echo "  destroy                - Eliminar cluster completamente"
	@echo ""
	@echo "🔄 Arquitectura desplegada:"
	@echo "  Flujo Principal: blockchain"
	@echo ""
	@echo "Ejemplos:"
	@echo "  make init              # Usa Kind por defecto"
	@echo "  make deploy            # Despliega arquitectura completa"
	@echo "  make status            # Ver estado de componentes"
	
init: ## Crear cluster local con Kind o Minikube
	@echo "🚀 Creando cluster local con $(PROVIDER)..."
ifeq ($(PROVIDER),kind)
	@if kind get clusters | grep -q $(CLUSTER_NAME); then \
		echo "⚠️  El cluster $(CLUSTER_NAME) ya existe"; \
	else \
		kind create cluster --name $(CLUSTER_NAME) --config=config/kind-config.yaml; \
	fi
	@echo "✅ Configurando kubectl context..."
	@kubectl config use-context kind-$(CLUSTER_NAME)
else ifeq ($(PROVIDER),minikube)
	@if minikube status -p $(CLUSTER_NAME) >/dev/null 2>&1; then \
		echo "⚠️  El cluster $(CLUSTER_NAME) ya existe"; \
	else \
		minikube start -p $(CLUSTER_NAME) \
			--driver=docker \
			--cpus=2 \
			--memory=6144 \
			--disk-size=20g \
			--kubernetes-version=stable \
			--addons=ingress,dashboard,metrics-server; \
	fi
	@echo "✅ Configurando kubectl context..."
	@kubectl config use-context $(CLUSTER_NAME)
else
	@echo "❌ PROVIDER debe ser 'kind' o 'minikube'"
	@exit 1
endif
	@echo "🎉 Cluster $(CLUSTER_NAME) con $(PROVIDER) listo!"

ifeq ($(PROVIDER),kind)
	@echo "🔨 Construyendo y cargando imágenes Docker en Kind..."
	@$(MAKE) -C ../services build-load-all
endif

	@echo "📦 Desplegando charts en el cluster..."
	@echo "🔧 Verificando que el cluster esté activo..."
ifeq ($(PROVIDER),kind)
	@kubectl cluster-info --context kind-$(CLUSTER_NAME) > /dev/null
else
	@kubectl cluster-info --context $(CLUSTER_NAME) > /dev/null
endif
	
	@echo "📋 Agregando repositorios de Helm..."
#	@helm repo add istio https://istio-release.storage.googleapis.com/charts || true
#	@helm repo add kedacore https://kedacore.github.io/charts || true
#	@helm repo update
	
	@echo "🚀 Desplegando Istio Base..."
	@helm upgrade --install istio-base ./istio/base \
		--namespace istio-system \
		--set defaultRevision=default \
		--create-namespace \
		--wait
	
	@echo "🔧 Desplegando Istiod..."
	@helm upgrade --install istiod ./istio/istiod \
		--namespace istio-system \
		--wait

	@echo "📦 AddOns no está instalado. Instalando AddOns..."; \
		kubectl apply -f https://raw.githubusercontent.com/istio/istio/release-1.27/samples/addons/prometheus.yaml; \
		kubectl apply -f https://raw.githubusercontent.com/istio/istio/release-1.27/samples/addons/jaeger.yaml; \
		kubectl apply -f https://raw.githubusercontent.com/istio/istio/release-1.27/samples/addons/grafana.yaml; \
		kubectl apply -f https://raw.githubusercontent.com/istio/istio/release-1.27/samples/addons/kiali.yaml; \
		echo "⏳ Esperando a que AddOns esté listo..."; \
		kubectl wait --for=condition=available --timeout=300s deployment/prometheus -n istio-system; \
		kubectl wait --for=condition=available --timeout=300s deployment/jaeger -n istio-system; \
		kubectl wait --for=condition=available --timeout=300s deployment/grafana -n istio-system; \
		kubectl wait --for=condition=available --timeout=300s deployment/kiali -n istio-system; \

	@echo "🌐 Desplegando Istio Gateway..."
#	@helm upgrade --install istio-gateway ./istio/gateway \
		--namespace istio-system \
		--set profile=demo \
		--wait
	
	@echo "⚡ Desplegando KEDA..."
	@helm upgrade --install keda ./keda \
		--namespace keda-system \
		--create-namespace \
		--wait
	
	@echo "✅ Todos los charts desplegados exitosamente!"

deploy:
	@kubectl create namespace medisupply --dry-run=client -o yaml | kubectl apply -f - || echo "✅ Namespace medisupply ya existe o fue creado"
	@kubectl label namespace medisupply istio-injection=enabled

	@kubectl create namespace medilogistic --dry-run=client -o yaml | kubectl apply -f - || echo "✅ Namespace medilogistic ya existe o fue creado"
	@kubectl label namespace medilogistic istio-injection=enabled

	@kubectl create namespace mediorder --dry-run=client -o yaml | kubectl apply -f - || echo "✅ Namespace mediorder ya existe o fue creado"
	@kubectl label namespace mediorder istio-injection=enabled

	@kubectl create namespace mediwarehouse --dry-run=client -o yaml | kubectl apply -f - || echo "✅ Namespace mediwarehouse ya existe o fue creado"
	@kubectl label namespace mediwarehouse istio-injection=enabled
    
	@echo ""
	@echo "🎉 ¡Despliegue completado exitosamente!"
	@echo "📋 Servicios desplegados:"
	@echo "  ✅ Blockchain"
	@echo ""

clean: ## Eliminar charts del cluster

	@echo "🗑️  Eliminando namespace medisupply..."
	@kubectl delete namespace medisupply --ignore-not-found=true
	@kubectl delete namespace medilogistic --ignore-not-found=true
	@kubectl delete namespace mediorder --ignore-not-found=true
	@kubectl delete namespace mediwarehouse --ignore-not-found=true
	@echo ""
	@echo "✅ Todos los charts eliminados del cluster!"
	@echo "📋 Servicios eliminados:"
	@echo "  ❌ Namespace medisupply"
	@echo "  ❌ Namespace medilogistic"
	@echo "  ❌ Namespace mediorder"
	@echo "  ❌ Namespace mediwarehouse"

destroy: ## Eliminar cluster completamente
	@echo "💥 Eliminando cluster $(CLUSTER_NAME)..."
ifeq ($(PROVIDER),kind)
	@if kind get clusters | grep -q $(CLUSTER_NAME); then \
		kind delete cluster --name $(CLUSTER_NAME); \
		echo "✅ Cluster $(CLUSTER_NAME) eliminado completamente!"; \
	else \
		echo "⚠️  El cluster $(CLUSTER_NAME) no existe"; \
	fi
else ifeq ($(PROVIDER),minikube)
	@if minikube status -p $(CLUSTER_NAME) >/dev/null 2>&1; then \
		minikube delete -p $(CLUSTER_NAME); \
		echo "✅ Cluster $(CLUSTER_NAME) eliminado completamente!"; \
	else \
		echo "⚠️  El cluster $(CLUSTER_NAME) no existe"; \
	fi
else
	@echo "❌ PROVIDER debe ser 'kind' o 'minikube'"
	@exit 1
endif


kiali: ## Abrir Kiali dashboard
	@echo "🔍 Abriendo Kiali dashboard..."
	@echo "📡 Configurando port-forward para Kiali..."
	@kubectl port-forward -n istio-system svc/kiali 20001:20001 &
	@sleep 3
	@echo "🌐 Kiali disponible en: http://localhost:20001"
	@echo "💡 Presiona Ctrl+C para detener el port-forward"

