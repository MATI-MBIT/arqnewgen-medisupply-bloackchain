package handlers

import (
	"AlchemyWebSocketMicro/models"
	"AlchemyWebSocketMicro/services"
	"encoding/json"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type WebSocketHandler struct {
	alchemyService *services.AlchemyService
	upgrader       websocket.Upgrader
}

func NewWebSocketHandler(alchemyService *services.AlchemyService) *WebSocketHandler {
	return &WebSocketHandler{
		alchemyService: alchemyService,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Permitir todas las conexiones en desarrollo
			},
		},
	}
}

func (h *WebSocketHandler) HealthCheck(c *gin.Context) {
	log.Printf("🏥 Health check solicitado")
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "AlchemyWebSocketMicro está funcionando correctamente",
		"service": "alchemy-websocket-micro",
	})
}

func (h *WebSocketHandler) HandleWebSocket(c *gin.Context) {
	contractAddress := c.Param("contractAddress")
	
	log.Printf("🔌 Nueva conexión WebSocket solicitada para contrato: %s", contractAddress)
	log.Printf("📍 Cliente IP: %s", c.ClientIP())
	log.Printf("🌐 User-Agent: %s", c.GetHeader("User-Agent"))

	// Validar dirección del contrato
	if contractAddress == "" {
		log.Printf("❌ Dirección de contrato vacía")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Contract address is required"})
		return
	}

	if len(contractAddress) != 42 || contractAddress[:2] != "0x" {
		log.Printf("❌ Dirección de contrato inválida: %s", contractAddress)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid contract address format"})
		return
	}

	// Upgrade a WebSocket
	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("❌ Error upgrading to WebSocket: %v", err)
		return
	}
	defer func() {
		log.Printf("🔌 Cerrando conexión WebSocket para: %s", contractAddress)
		h.alchemyService.UnsubscribeClient(conn)
		conn.Close()
	}()

	log.Printf("✅ WebSocket connection establecida para: %s", contractAddress)

	// Suscribirse al contrato en Alchemy
	if err := h.alchemyService.SubscribeToContract(contractAddress, conn); err != nil {
		log.Printf("❌ Error suscribiéndose al contrato %s: %v", contractAddress, err)
		
		errorMsg := models.WebSocketMessage{
			Type:         "error",
			ContractAddr: contractAddress,
			Data:         map[string]string{"error": err.Error()},
		}
		
		if msgBytes, marshalErr := json.Marshal(errorMsg); marshalErr == nil {
			conn.WriteMessage(websocket.TextMessage, msgBytes)
		}
		return
	}

	// Enviar mensaje de confirmación
	confirmMsg := models.WebSocketMessage{
		Type:         "connected",
		ContractAddr: contractAddress,
		Data:         map[string]string{"status": "monitoring started"},
	}
	
	if msgBytes, err := json.Marshal(confirmMsg); err == nil {
		conn.WriteMessage(websocket.TextMessage, msgBytes)
		log.Printf("📤 Mensaje de confirmación enviado para: %s", contractAddress)
	}

	// Mantener conexión viva y manejar mensajes del cliente
	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("❌ Error inesperado en WebSocket para %s: %v", contractAddress, err)
			} else {
				log.Printf("🔌 Cliente desconectado normalmente para: %s", contractAddress)
			}
			break
		}

		log.Printf("📨 Mensaje recibido del cliente %s (tipo: %d): %s", contractAddress, messageType, string(message))

		// Echo del mensaje (opcional, para debugging)
		if messageType == websocket.TextMessage {
			echoMsg := models.WebSocketMessage{
				Type:         "echo",
				ContractAddr: contractAddress,
				Data:         string(message),
			}
			
			if msgBytes, err := json.Marshal(echoMsg); err == nil {
				conn.WriteMessage(websocket.TextMessage, msgBytes)
				log.Printf("📤 Echo enviado para: %s", contractAddress)
			}
		}
	}
}

func (h *WebSocketHandler) GetMonitorStatus(c *gin.Context) {
	log.Printf("📊 Estado de monitoreo solicitado")
	
	statuses := h.alchemyService.GetActiveSubscriptions()
	
	log.Printf("📈 Suscripciones activas: %d", len(statuses))
	for _, status := range statuses {
		log.Printf("   - Contrato: %s, Activo: %t, ID: %s", 
			status.ContractAddress, status.IsActive, status.SubscriptionID)
	}

	c.JSON(http.StatusOK, gin.H{
		"success":       true,
		"message":       "Estado de monitoreo obtenido exitosamente",
		"activeMonitors": len(statuses),
		"subscriptions": statuses,
	})
}

func (h *WebSocketHandler) StartMonitoring(c *gin.Context) {
	contractAddress := c.Param("contractAddress")
	
	log.Printf("🚀 Solicitud de inicio de monitoreo para: %s", contractAddress)

	if contractAddress == "" {
		log.Printf("❌ Dirección de contrato vacía en start monitoring")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Contract address is required"})
		return
	}

	// Esta función es principalmente informativa ya que el monitoreo real
	// se inicia cuando se conecta un cliente WebSocket
	log.Printf("ℹ️ Para iniciar monitoreo real, conecte via WebSocket a: ws://localhost:8081/ws/monitor/%s", contractAddress)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Para iniciar monitoreo, conecte via WebSocket",
		"websocketUrl": "/ws/monitor/" + contractAddress,
		"contractAddress": contractAddress,
	})
}

func (h *WebSocketHandler) StopMonitoring(c *gin.Context) {
	contractAddress := c.Param("contractAddress")
	
	log.Printf("🛑 Solicitud de detener monitoreo para: %s", contractAddress)

	if contractAddress == "" {
		log.Printf("❌ Dirección de contrato vacía en stop monitoring")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Contract address is required"})
		return
	}

	// El monitoreo se detiene automáticamente cuando se desconectan todos los clientes WebSocket
	log.Printf("ℹ️ El monitoreo se detiene automáticamente cuando se desconectan todos los clientes WebSocket")

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "El monitoreo se detiene cuando se desconectan todos los clientes",
		"contractAddress": contractAddress,
	})
}