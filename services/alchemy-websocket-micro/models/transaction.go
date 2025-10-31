package models

import "encoding/json"

// AlchemyRequest representa la estructura de request a Alchemy
type AlchemyRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
	ID      int         `json:"id"`
}

// AlchemyResponse representa la respuesta de Alchemy
type AlchemyResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int             `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *AlchemyError   `json:"error,omitempty"`
}

// AlchemyError representa un error de Alchemy
type AlchemyError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// SubscriptionParams parámetros para la suscripción
type SubscriptionParams struct {
	Addresses      []AddressFilter `json:"addresses"`
	IncludeRemoved bool            `json:"includeRemoved"`
	HashesOnly     bool            `json:"hashesOnly"`
}

// AddressFilter filtro por dirección
type AddressFilter struct {
	To string `json:"to"`
}

// TransactionNotification notificación de transacción
type TransactionNotification struct {
	JSONRPC string `json:"jsonrpc"`
	Method  string `json:"method"`
	Params  struct {
		Subscription string          `json:"subscription"`
		Result       json.RawMessage `json:"result"`
	} `json:"params"`
}

// WebSocketMessage mensaje para enviar a clientes WebSocket
type WebSocketMessage struct {
	Type         string      `json:"type"`
	ContractAddr string      `json:"contractAddress"`
	Data         interface{} `json:"data"`
	Timestamp    int64       `json:"timestamp"`
}

// MonitorStatus estado del monitoreo
type MonitorStatus struct {
	ContractAddress string `json:"contractAddress"`
	IsActive        bool   `json:"isActive"`
	SubscriptionID  string `json:"subscriptionId,omitempty"`
	ConnectedAt     int64  `json:"connectedAt"`
}