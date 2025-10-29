# AlchemyWebSocketMicro Makefile

.PHONY: help build run test clean docker-build docker-run docker-stop deps fmt vet

# Variables
APP_NAME=alchemy-websocket-micro
DOCKER_IMAGE=$(APP_NAME)
PORT=8081

help: ## Mostrar ayuda
	@echo "Comandos disponibles:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

deps: ## Instalar dependencias
	@echo "📦 Instalando dependencias..."
	go mod download
	go mod tidy

fmt: ## Formatear código
	@echo "🎨 Formateando código..."
	go fmt ./...

vet: ## Verificar código
	@echo "🔍 Verificando código..."
	go vet ./...

build: deps fmt vet ## Build de la aplicación
	@echo "🔨 Building aplicación..."
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

run: ## Ejecutar aplicación localmente
	@echo "🚀 Ejecutando aplicación..."
	go run main.go

test: ## Ejecutar tests
	@echo "🧪 Ejecutando tests..."
	go test -v ./...

clean: ## Limpiar archivos generados
	@echo "🧹 Limpiando archivos..."
	rm -f main
	go clean

docker-build: ## Build imagen Docker
	@echo "🐳 Building imagen Docker..."
	docker build -t $(DOCKER_IMAGE) .

docker-run: docker-build ## Ejecutar con Docker
	@echo "🐳 Ejecutando con Docker..."
	docker run -p $(PORT):$(PORT) \
		-e ALCHEMY_API_KEY=${ALCHEMY_API_KEY} \
		-e ALCHEMY_WS_URL=${ALCHEMY_WS_URL} \
		-e PORT=$(PORT) \
		--name $(APP_NAME) \
		$(DOCKER_IMAGE)

docker-stop: ## Detener contenedor Docker
	@echo "🛑 Deteniendo contenedor..."
	docker stop $(APP_NAME) || true
	docker rm $(APP_NAME) || true

docker-compose-up: ## Ejecutar con docker-compose
	@echo "🐳 Ejecutando con docker-compose..."
	docker-compose up --build

docker-compose-down: ## Detener docker-compose
	@echo "🛑 Deteniendo docker-compose..."
	docker-compose down

logs: ## Ver logs del contenedor
	@echo "📋 Mostrando logs..."
	docker logs -f $(APP_NAME)

health: ## Verificar health del servicio
	@echo "🏥 Verificando health..."
	curl -s http://localhost:$(PORT)/api/v1/health | jq .

status: ## Ver estado de monitoreo
	@echo "📊 Verificando estado..."
	curl -s http://localhost:$(PORT)/api/v1/monitor/status | jq .

wscat-test: ## Probar WebSocket con wscat (requiere wscat instalado)
	@echo "🔌 Probando WebSocket..."
	@echo "Conectando a ws://localhost:$(PORT)/ws/monitor/0x9d70c560cE7D6EDAaf4562E980136D21Fd0fbdc9"
	wscat -c ws://localhost:$(PORT)/ws/monitor/0x9d70c560cE7D6EDAaf4562E980136D21Fd0fbdc9

install-wscat: ## Instalar wscat para testing
	@echo "📦 Instalando wscat..."
	npm install -g wscat

dev: ## Modo desarrollo con auto-reload (requiere air)
	@echo "🔄 Iniciando en modo desarrollo..."
	air

install-air: ## Instalar air para auto-reload
	@echo "📦 Instalando air..."
	go install github.com/cosmtrek/air@latest

setup: deps install-wscat ## Setup completo para desarrollo
	@echo "✅ Setup completo"
	@echo "Comandos útiles:"
	@echo "  make run          - Ejecutar aplicación"
	@echo "  make wscat-test   - Probar WebSocket"
	@echo "  make health       - Verificar health"
	@echo "  make status       - Ver estado"