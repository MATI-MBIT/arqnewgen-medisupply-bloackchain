package main

import (
	"AlchemyWebSocketMicro/config"
	"AlchemyWebSocketMicro/handlers"
	"AlchemyWebSocketMicro/services"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	log.Printf("🚀 Iniciando AlchemyWebSocketMicro...")

	// Cargar configuración
	cfg := config.LoadConfig()
	log.Printf("⚙️ Configuración cargada - Puerto: %s", cfg.Port)
	log.Printf("🔗 Alchemy WebSocket URL: %s", cfg.AlchemyWSURL)

	// Inicializar servicio de Alchemy
	alchemyService := services.NewAlchemyService(cfg.AlchemyWSURL, cfg.AlchemyAPIKey)
	
	log.Printf("🔌 Iniciando conexión con Alchemy...")
	if err := alchemyService.Start(); err != nil {
		log.Fatalf("❌ Error iniciando servicio de Alchemy: %v", err)
	}
	log.Printf("✅ Servicio de Alchemy iniciado exitosamente")

	// Inicializar handlers
	wsHandler := handlers.NewWebSocketHandler(alchemyService)

	// Configurar Gin
	r := gin.Default()

	// Middleware para logging de requests
	r.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("🌐 %s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
			param.ClientIP,
			param.TimeStamp.Format("2006/01/02 - 15:04:05"),
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
		)
	}))

	// Middleware para CORS
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, Upgrade, Connection, Sec-WebSocket-Key, Sec-WebSocket-Version, Sec-WebSocket-Protocol")

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
		api.GET("/health", wsHandler.HealthCheck)

		// Rutas de monitoreo
		monitor := api.Group("/monitor")
		{
			monitor.POST("/start/:contractAddress", wsHandler.StartMonitoring)
			monitor.POST("/stop/:contractAddress", wsHandler.StopMonitoring)
			monitor.GET("/status", wsHandler.GetMonitorStatus)
		}
	}

	// Rutas WebSocket
	ws := r.Group("/ws")
	{
		ws.GET("/monitor/:contractAddress", wsHandler.HandleWebSocket)
	}

	// Ruta de información
	r.GET("/", func(c *gin.Context) {
		log.Printf("📋 Información del servicio solicitada")
		c.JSON(200, gin.H{
			"service":     "AlchemyWebSocketMicro",
			"version":     "1.0.0",
			"description": "Microservicio WebSocket para monitoreo de transacciones Ethereum via Alchemy",
			"endpoints": gin.H{
				"health":    "/api/v1/health",
				"websocket": "/ws/monitor/{contractAddress}",
				"status":    "/api/v1/monitor/status",
			},
			"usage": gin.H{
				"websocket": "ws://localhost:" + cfg.Port + "/ws/monitor/0x1234567890123456789012345678901234567890",
				"example":   "wscat -c ws://localhost:" + cfg.Port + "/ws/monitor/0x9d70c560cE7D6EDAaf4562E980136D21Fd0fbdc9",
			},
		})
	})

	// Iniciar servidor
	log.Printf("🌐 AlchemyWebSocketMicro iniciando en puerto %s", cfg.Port)
	log.Printf("📡 WebSocket endpoint: ws://localhost:%s/ws/monitor/{contractAddress}", cfg.Port)
	log.Printf("🔗 API REST: http://localhost:%s/api/v1/", cfg.Port)
	log.Printf("📋 Información: http://localhost:%s/", cfg.Port)
	
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("❌ Error iniciando servidor: %v", err)
	}
}