package services

import (
	"AlchemyWebSocketMicro/models"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type AlchemyService struct {
	wsURL       string
	apiKey      string
	conn        *websocket.Conn
	subscribers map[string]*Subscription
	mu          sync.RWMutex
	reconnectCh chan bool
}

type Subscription struct {
	ContractAddress string
	SubscriptionID  string
	Clients         map[*websocket.Conn]bool
	ClientsMu       sync.RWMutex
}

func NewAlchemyService(wsURL, apiKey string) *AlchemyService {
	return &AlchemyService{
		wsURL:       fmt.Sprintf("%s/%s", wsURL, apiKey),
		apiKey:      apiKey,
		subscribers: make(map[string]*Subscription),
		reconnectCh: make(chan bool, 1),
	}
}

func (a *AlchemyService) Start() error {
	log.Printf("🚀 Iniciando AlchemyService con URL: %s", a.wsURL)
	
	if err := a.connect(); err != nil {
		return fmt.Errorf("error conectando a Alchemy: %v", err)
	}

	go a.handleReconnection()
	go a.readMessages()

	return nil
}

func (a *AlchemyService) connect() error {
	log.Printf("🔌 Conectando a Alchemy WebSocket...")
	
	conn, _, err := websocket.DefaultDialer.Dial(a.wsURL, nil)
	if err != nil {
		return fmt.Errorf("error en dial: %v", err)
	}

	a.conn = conn
	log.Printf("✅ Conectado exitosamente a Alchemy WebSocket")
	return nil
}

func (a *AlchemyService) handleReconnection() {
	for range a.reconnectCh {
		log.Printf("🔄 Intentando reconectar a Alchemy...")
		
		for {
			if err := a.connect(); err != nil {
				log.Printf("❌ Error en reconexión: %v. Reintentando en 5s...", err)
				time.Sleep(5 * time.Second)
				continue
			}

			log.Printf("✅ Reconectado exitosamente")
			
			// Re-suscribir a todos los contratos activos
			a.mu.RLock()
			for contractAddr := range a.subscribers {
				go a.subscribeToContract(contractAddr)
			}
			a.mu.RUnlock()
			break
		}
	}
}

func (a *AlchemyService) readMessages() {
	defer func() {
		if a.conn != nil {
			a.conn.Close()
		}
	}()

	for {
		if a.conn == nil {
			log.Printf("⚠️ Conexión WebSocket es nil, solicitando reconexión...")
			select {
			case a.reconnectCh <- true:
			default:
			}
			return
		}

		_, message, err := a.conn.ReadMessage()
		if err != nil {
			log.Printf("❌ Error leyendo mensaje de Alchemy: %v", err)
			select {
			case a.reconnectCh <- true:
			default:
			}
			return
		}

		log.Printf("📨 Mensaje recibido de Alchemy: %s", string(message))
		a.handleAlchemyMessage(message)
	}
}

func (a *AlchemyService) handleAlchemyMessage(message []byte) {
	// Intentar parsear como respuesta de suscripción
	var response models.AlchemyResponse
	if err := json.Unmarshal(message, &response); err == nil && response.Result != nil {
		log.Printf("📋 Respuesta de suscripción recibida - ID: %d", response.ID)
		a.handleSubscriptionResponse(&response)
		return
	}

	// Intentar parsear como notificación de transacción
	var notification models.TransactionNotification
	if err := json.Unmarshal(message, &notification); err == nil && notification.Method == "eth_subscription" {
		log.Printf("🔔 Notificación de transacción recibida - Subscription: %s", notification.Params.Subscription)
		a.handleTransactionNotification(&notification)
		return
	}

	log.Printf("⚠️ Mensaje no reconocido de Alchemy: %s", string(message))
}

func (a *AlchemyService) handleSubscriptionResponse(response *models.AlchemyResponse) {
	var subscriptionID string
	if err := json.Unmarshal(response.Result, &subscriptionID); err != nil {
		log.Printf("❌ Error parseando subscription ID: %v", err)
		return
	}

	log.Printf("✅ Suscripción creada exitosamente - ID: %s", subscriptionID)
	
	// Actualizar el subscription ID en la suscripción correspondiente
	a.mu.Lock()
	for _, sub := range a.subscribers {
		if sub.SubscriptionID == "" {
			sub.SubscriptionID = subscriptionID
			log.Printf("🔗 Subscription ID %s asignado a contrato %s", subscriptionID, sub.ContractAddress)
			break
		}
	}
	a.mu.Unlock()
}

func (a *AlchemyService) handleTransactionNotification(notification *models.TransactionNotification) {
	log.Printf("💰 Procesando transacción - Subscription: %s", notification.Params.Subscription)
	log.Printf("📄 Datos de transacción: %s", string(notification.Params.Result))

	// Encontrar la suscripción correspondiente
	a.mu.RLock()
	var targetSub *Subscription
	for _, sub := range a.subscribers {
		if sub.SubscriptionID == notification.Params.Subscription {
			targetSub = sub
			break
		}
	}
	a.mu.RUnlock()

	if targetSub == nil {
		log.Printf("⚠️ No se encontró suscripción para ID: %s", notification.Params.Subscription)
		return
	}

	log.Printf("🎯 Enviando transacción a clientes del contrato: %s", targetSub.ContractAddress)

	// Crear mensaje para clientes WebSocket
	wsMessage := models.WebSocketMessage{
		Type:         "transaction",
		ContractAddr: targetSub.ContractAddress,
		Data:         json.RawMessage(notification.Params.Result),
		Timestamp:    time.Now().Unix(),
	}

	messageBytes, err := json.Marshal(wsMessage)
	if err != nil {
		log.Printf("❌ Error serializando mensaje: %v", err)
		return
	}

	// Enviar a todos los clientes conectados
	targetSub.ClientsMu.RLock()
	clientCount := len(targetSub.Clients)
	log.Printf("📤 Enviando a %d clientes conectados", clientCount)
	
	for client := range targetSub.Clients {
		if err := client.WriteMessage(websocket.TextMessage, messageBytes); err != nil {
			log.Printf("❌ Error enviando mensaje a cliente: %v", err)
			// Remover cliente desconectado
			delete(targetSub.Clients, client)
			client.Close()
		} else {
			log.Printf("✅ Mensaje enviado exitosamente a cliente")
		}
	}
	targetSub.ClientsMu.RUnlock()
}

func (a *AlchemyService) SubscribeToContract(contractAddress string, client *websocket.Conn) error {
	log.Printf("🔔 Iniciando suscripción para contrato: %s", contractAddress)

	a.mu.Lock()
	defer a.mu.Unlock()

	// Verificar si ya existe suscripción para este contrato
	if sub, exists := a.subscribers[contractAddress]; exists {
		log.Printf("📌 Suscripción existente encontrada para: %s", contractAddress)
		sub.ClientsMu.Lock()
		sub.Clients[client] = true
		sub.ClientsMu.Unlock()
		log.Printf("👥 Cliente agregado a suscripción existente. Total clientes: %d", len(sub.Clients))
		return nil
	}

	// Crear nueva suscripción
	subscription := &Subscription{
		ContractAddress: contractAddress,
		Clients:         make(map[*websocket.Conn]bool),
	}
	subscription.Clients[client] = true
	a.subscribers[contractAddress] = subscription

	log.Printf("🆕 Nueva suscripción creada para: %s", contractAddress)

	// Suscribirse a Alchemy
	return a.subscribeToContract(contractAddress)
}

func (a *AlchemyService) subscribeToContract(contractAddress string) error {
	log.Printf("📡 Enviando suscripción a Alchemy para: %s", contractAddress)

	request := models.AlchemyRequest{
		JSONRPC: "2.0",
		Method:  "eth_subscribe",
		Params: []interface{}{
			"alchemy_minedTransactions",
			models.SubscriptionParams{
				Addresses: []models.AddressFilter{
					{To: contractAddress},
				},
				IncludeRemoved: false,
				HashesOnly:     false,
			},
		},
		ID: int(time.Now().Unix()),
	}

	requestBytes, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("error serializando request: %v", err)
	}

	log.Printf("📤 Enviando request a Alchemy: %s", string(requestBytes))

	if a.conn == nil {
		return fmt.Errorf("conexión WebSocket no disponible")
	}

	if err := a.conn.WriteMessage(websocket.TextMessage, requestBytes); err != nil {
		return fmt.Errorf("error enviando mensaje a Alchemy: %v", err)
	}

	log.Printf("✅ Request enviado exitosamente a Alchemy")
	return nil
}

func (a *AlchemyService) UnsubscribeClient(client *websocket.Conn) {
	log.Printf("🔌 Desconectando cliente...")

	a.mu.Lock()
	defer a.mu.Unlock()

	for contractAddr, sub := range a.subscribers {
		sub.ClientsMu.Lock()
		if _, exists := sub.Clients[client]; exists {
			delete(sub.Clients, client)
			log.Printf("👋 Cliente removido de suscripción: %s", contractAddr)
			
			// Si no quedan clientes, remover la suscripción
			if len(sub.Clients) == 0 {
				delete(a.subscribers, contractAddr)
				log.Printf("🗑️ Suscripción removida para: %s (sin clientes)", contractAddr)
			}
		}
		sub.ClientsMu.Unlock()
	}
}

func (a *AlchemyService) GetActiveSubscriptions() []models.MonitorStatus {
	a.mu.RLock()
	defer a.mu.RUnlock()

	var statuses []models.MonitorStatus
	for contractAddr, sub := range a.subscribers {
		sub.ClientsMu.RLock()
		status := models.MonitorStatus{
			ContractAddress: contractAddr,
			IsActive:        len(sub.Clients) > 0,
			SubscriptionID:  sub.SubscriptionID,
			ConnectedAt:     time.Now().Unix(),
		}
		sub.ClientsMu.RUnlock()
		statuses = append(statuses, status)
	}

	return statuses
}