package main

import (
	"CrearLoteMicro/config"
	"CrearLoteMicro/handlers"
	"CrearLoteMicro/services"
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	// Cargar configuración
	cfg := config.LoadConfig()

	// Inicializar servicio de blockchain
	blockchainService, err := services.NewBlockchainService(cfg.SepoliaRPC, cfg.ChainID)
	if err != nil {
		log.Fatalf("Error inicializando servicio de blockchain: %v", err)
	}
	// Inicializar servicio de blockchainSocket
	blockchainWebsocketService := services.NewBlockchainWebsocketService(cfg.SepoliaWS, blockchainService)

	// Inicializar handlers
	loteHandler := handlers.NewLoteHandler(blockchainService, blockchainWebsocketService)

	// Configurar Gin
	r := gin.Default()

	// Middleware para CORS
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Rutas de la API
	api := r.Group("/api/v1")
	{
		// Health check
		api.GET("/health", loteHandler.HealthCheck)

		// Rutas de lote
		lote := api.Group("/lote")
		{
			lote.POST("/crear", loteHandler.CrearLote)
			lote.POST("/temperatura", loteHandler.RegistrarTemperatura)
			lote.POST("/transferir", loteHandler.TransferirCustodia)
			lote.GET("/info/:contractAddress", loteHandler.ObtenerLote)
			lote.GET("/cadena/:contractAddress", loteHandler.ObtenerCadenaBlockchain)
		}

		// Rutas de utilidades
		utils := api.Group("/utils")
		{
			utils.POST("/decode", loteHandler.DecodificarInputData)
			utils.GET("/decode/specific", loteHandler.DecodificarInputDataEspecifico)
			utils.GET("/signatures", loteHandler.ObtenerSignaturesFunciones)
		}

		// Rutas de diagnóstico
		debug := api.Group("/debug")
		{
			debug.GET("/conexion", loteHandler.VerificarConexion)
			debug.GET("/contrato/:contractAddress", loteHandler.DiagnosticarContrato)
		}
	}

	// Iniciar servidor
	log.Printf("CrearLoteMicro iniciando en puerto %s", cfg.Port)
	log.Printf("Conectado a Sepolia RPC: %s", cfg.SepoliaRPC)
	log.Printf("Conectado a Sepolia WS: %s", cfg.SepoliaWS)

	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Error iniciando servidor: %v", err)
	}
}
