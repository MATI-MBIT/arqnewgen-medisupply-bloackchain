# Makefile para CrearLoteMicro
# Gestión de assets de contratos y operaciones de desarrollo

# Configuración
HARDHAT_ARTIFACT := ../../smartcontract/lotetracing/artifacts/contracts/LoteTracing.sol/LoteTracing.json
ASSETS_DIR := ./assets/contracts
ABI_FILE := $(ASSETS_DIR)/LoteTracing.abi.json
BYTECODE_FILE := $(ASSETS_DIR)/LoteTracing.bytecode
INFO_FILE := $(ASSETS_DIR)/contract_info.json
CURRENT_DATE := $(shell date +"%Y-%m-%d")

# Colores para output
RED := \033[0;31m
GREEN := \033[0;32m
YELLOW := \033[1;33m
BLUE := \033[0;34m
NC := \033[0m # No Color

.PHONY: help update-contract-assets validate-assets build test clean run docker-build docker-run

help: ## Mostrar ayuda
	@echo "$(YELLOW)CrearLoteMicro - Comandos disponibles:$(NC)"
	@echo ""
	@echo "$(BLUE)📄 Gestión de Assets:$(NC)"
	@echo "  update-contract-assets  - Actualizar assets del contrato desde Hardhat"
	@echo "  validate-assets         - Validar integridad de los assets"
	@echo ""
	@echo "$(BLUE)🔨 Desarrollo:$(NC)"
	@echo "  build                   - Compilar el microservicio"
	@echo "  test                    - Ejecutar tests"
	@echo "  run                     - Ejecutar el microservicio localmente"
	@echo "  clean                   - Limpiar archivos temporales"
	@echo ""
	@echo "$(BLUE)🐳 Docker:$(NC)"
	@echo "  docker-build            - Construir imagen Docker"
	@echo "  docker-run              - Ejecutar contenedor Docker"
	@echo ""
	@echo "$(BLUE)📋 Información:$(NC)"
	@echo "  show-contract-info      - Mostrar información del contrato actual"
	@echo ""

update-contract-assets: ## Actualizar assets del contrato desde Hardhat
	@echo "$(YELLOW)🔄 Actualizando assets del contrato LoteTracing...$(NC)"
	@if [ ! -f "$(HARDHAT_ARTIFACT)" ]; then \
		echo "$(RED)❌ Error: No se encontró el artifact de Hardhat en $(HARDHAT_ARTIFACT)$(NC)"; \
		echo "$(YELLOW)💡 Asegúrate de que el contrato esté compilado en Hardhat$(NC)"; \
		exit 1; \
	fi
	@mkdir -p $(ASSETS_DIR)
	@echo "$(YELLOW)📄 Extrayendo ABI...$(NC)"
	@jq '.abi' "$(HARDHAT_ARTIFACT)" > "$(ABI_FILE)"
	@echo "$(YELLOW)📄 Extrayendo Bytecode...$(NC)"
	@jq -r '.bytecode' "$(HARDHAT_ARTIFACT)" > "$(BYTECODE_FILE)"
	@echo "$(YELLOW)📄 Actualizando información del contrato...$(NC)"
	@BYTECODE_CONTENT=$$(cat "$(BYTECODE_FILE)"); \
	CONTRACT_HASH="0x$$(echo "$$BYTECODE_CONTENT" | tail -c 65 | head -c 64)"; \
	jq -n \
		--arg name "LoteTracing" \
		--arg version "1.0.0" \
		--arg compiler "solc-0.8.28" \
		--arg date "$(CURRENT_DATE)" \
		--arg source "smartcontract/lotetracing/artifacts/contracts/LoteTracing.sol/LoteTracing.json" \
		--arg hash "$$CONTRACT_HASH" \
		--arg desc "Smart contract para trazabilidad de lotes con control de temperatura" \
		'{ \
			contractName: $$name, \
			version: $$version, \
			compiler: $$compiler, \
			lastUpdated: $$date, \
			source: $$source, \
			hash: $$hash, \
			description: $$desc, \
			features: ["Registro de temperaturas", "Control de cadena de frío", "Transferencia de custodia", "Detección automática de compromiso"], \
			events: ["LoteCreado", "CustodiaTransferida", "LoteComprometido"], \
			functions: ["registrarTemperatura", "transferirCustodia", "comprometido", "fabricante", "loteId", "propietarioActual", "tempRegMaxima", "tempRegMinima", "temperaturaMaxima", "temperaturaMinima"] \
		}' > "$(INFO_FILE)"
	@echo "$(GREEN)✅ Assets actualizados exitosamente:$(NC)"
	@echo "   📄 ABI: $(ABI_FILE)"
	@echo "   📄 Bytecode: $(BYTECODE_FILE)"
	@echo "   📄 Info: $(INFO_FILE)"
	@$(MAKE) validate-assets

validate-assets: ## Validar integridad de los assets
	@echo "$(YELLOW)🔍 Validando assets del contrato...$(NC)"
	@if [ ! -f "$(ABI_FILE)" ]; then \
		echo "$(RED)❌ Error: ABI no encontrado en $(ABI_FILE)$(NC)"; \
		exit 1; \
	fi
	@if [ ! -f "$(BYTECODE_FILE)" ]; then \
		echo "$(RED)❌ Error: Bytecode no encontrado en $(BYTECODE_FILE)$(NC)"; \
		exit 1; \
	fi
	@if [ ! -f "$(INFO_FILE)" ]; then \
		echo "$(RED)❌ Error: Info del contrato no encontrada en $(INFO_FILE)$(NC)"; \
		exit 1; \
	fi
	@echo "$(YELLOW)🔍 Verificando formato del bytecode...$(NC)"
	@BYTECODE_CONTENT=$$(cat "$(BYTECODE_FILE)"); \
	if [ "$${BYTECODE_CONTENT#0x}" = "$$BYTECODE_CONTENT" ]; then \
		echo "$(RED)❌ Error: Bytecode no tiene formato hexadecimal válido$(NC)"; \
		exit 1; \
	fi
	@echo "$(YELLOW)🔍 Verificando compilación Go...$(NC)"
	@if ! go build -o /tmp/test_contracts ./assets/contracts/loader.go 2>/dev/null; then \
		echo "$(RED)❌ Error: El módulo de assets no compila$(NC)"; \
		exit 1; \
	fi
	@rm -f /tmp/test_contracts
	@if ! go build -o /tmp/test_service ./services/blockchain_service.go 2>/dev/null; then \
		echo "$(RED)❌ Error: El servicio blockchain no compila$(NC)"; \
		exit 1; \
	fi
	@rm -f /tmp/test_service
	@echo "$(GREEN)✅ Todos los assets son válidos$(NC)"

show-contract-info: ## Mostrar información del contrato actual
	@echo "$(YELLOW)📋 Información del contrato actual:$(NC)"
	@if [ -f "$(INFO_FILE)" ]; then \
		jq . "$(INFO_FILE)"; \
	else \
		echo "$(RED)❌ No se encontró información del contrato$(NC)"; \
		echo "$(YELLOW)💡 Ejecuta 'make update-contract-assets' primero$(NC)"; \
	fi

build: ## Compilar el microservicio
	@echo "$(YELLOW)🔨 Compilando CrearLoteMicro...$(NC)"
	@go build -o bin/CrearLoteMicro .
	@echo "$(GREEN)✅ Compilación exitosa: bin/CrearLoteMicro$(NC)"

test: ## Ejecutar tests
	@echo "$(YELLOW)🧪 Ejecutando tests...$(NC)"
	@go test ./... -v
	@echo "$(GREEN)✅ Tests completados$(NC)"

run: ## Ejecutar el microservicio localmente
	@echo "$(YELLOW)🚀 Ejecutando CrearLoteMicro...$(NC)"
	@go run .

clean: ## Limpiar archivos temporales
	@echo "$(YELLOW)🧹 Limpiando archivos temporales...$(NC)"
	@rm -rf bin/
	@rm -f /tmp/test_*
	@go clean
	@echo "$(GREEN)✅ Limpieza completada$(NC)"

docker-build: ## Construir imagen Docker
	@echo "$(YELLOW)🐳 Construyendo imagen Docker...$(NC)"
	@docker build -t crearlote-micro:latest .
	@echo "$(GREEN)✅ Imagen Docker construida: crearlote-micro:latest$(NC)"

docker-run: ## Ejecutar contenedor Docker
	@echo "$(YELLOW)🐳 Ejecutando contenedor Docker...$(NC)"
	@docker run -p 8080:8080 crearlote-micro:latest

# Targets de desarrollo rápido
dev-setup: update-contract-assets build ## Configuración completa para desarrollo
	@echo "$(GREEN)🎉 Configuración de desarrollo completada$(NC)"

dev-update: update-contract-assets build ## Actualizar assets y recompilar
	@echo "$(GREEN)🔄 Actualización de desarrollo completada$(NC)"

# Verificar dependencias
check-deps: ## Verificar dependencias necesarias
	@echo "$(YELLOW)🔍 Verificando dependencias...$(NC)"
	@command -v jq >/dev/null 2>&1 || { echo "$(RED)❌ jq no está instalado$(NC)"; exit 1; }
	@command -v go >/dev/null 2>&1 || { echo "$(RED)❌ Go no está instalado$(NC)"; exit 1; }
	@command -v docker >/dev/null 2>&1 || { echo "$(RED)❌ Docker no está instalado$(NC)"; exit 1; }
	@echo "$(GREEN)✅ Todas las dependencias están disponibles$(NC)"