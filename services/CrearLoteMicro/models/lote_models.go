package models

// CrearLoteRequest representa la solicitud para crear un nuevo lote
type CrearLoteRequest struct {
	LoteID           string `json:"loteId" binding:"required"`
	TemperaturaMin   int8   `json:"temperaturaMin" binding:"required"`
	TemperaturaMax   int8   `json:"temperaturaMax" binding:"required"`
	WalletAddress    string `json:"walletAddress" binding:"required"`
	PrivateKey       string `json:"privateKey" binding:"required"`
}

// RegistrarTemperaturaRequest representa la solicitud para registrar temperatura
type RegistrarTemperaturaRequest struct {
	ContractAddress string `json:"contractAddress" binding:"required"`
	TempMin         int8   `json:"tempMin" binding:"required"`
	TempMax         int8   `json:"tempMax" binding:"required"`
	WalletAddress   string `json:"walletAddress" binding:"required"`
	PrivateKey      string `json:"privateKey" binding:"required"`
}

// TransferirCustodiaRequest representa la solicitud para transferir custodia
type TransferirCustodiaRequest struct {
	ContractAddress   string `json:"contractAddress" binding:"required"`
	NuevoPropietario  string `json:"nuevoPropietario" binding:"required"`
	WalletAddress     string `json:"walletAddress" binding:"required"`
	PrivateKey        string `json:"privateKey" binding:"required"`
}

// CrearNuevoLoteRequest representa la solicitud para crear un nuevo lote en un contrato existente
type CrearNuevoLoteRequest struct {
	ContractAddress string `json:"contractAddress" binding:"required"`
	LoteID          string `json:"loteId" binding:"required"`
	TemperaturaMin  int8   `json:"temperaturaMin" binding:"required"`
	TemperaturaMax  int8   `json:"temperaturaMax" binding:"required"`
	PrivateKey      string `json:"privateKey" binding:"required"`
}

// Response representa una respuesta genérica de la API
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	TxHash  string      `json:"txHash,omitempty"`
}

// ContractDeployResponse representa la respuesta al desplegar un contrato
type ContractDeployResponse struct {
	ContractAddress string `json:"contractAddress"`
	TxHash          string `json:"txHash"`
	LoteID          string `json:"loteId"`
}

// ObtenerLoteRequest representa la solicitud para obtener información de un lote
type ObtenerLoteRequest struct {
	ContractAddress string `json:"contractAddress" binding:"required"`
}

// LoteInfoResponse representa la información completa de un lote
type LoteInfoResponse struct {
	LoteID             string `json:"loteId"`
	Fabricante         string `json:"fabricante"`
	PropietarioActual  string `json:"propietarioActual"`
	TemperaturaMinima  int8   `json:"temperaturaMinima"`
	TemperaturaMaxima  int8   `json:"temperaturaMaxima"`
	TempRegMinima      int8   `json:"tempRegMinima"`
	TempRegMaxima      int8   `json:"tempRegMaxima"`
	Comprometido       bool   `json:"comprometido"`
	ContractAddress    string `json:"contractAddress"`
}

// EventoBlockchain representa un evento individual en la cadena
type EventoBlockchain struct {
	TipoEvento      string                 `json:"tipoEvento"`
	BlockNumber     uint64                 `json:"blockNumber"`
	TxHash          string                 `json:"txHash"`
	Timestamp       uint64                 `json:"timestamp"`
	Datos           map[string]interface{} `json:"datos"`
}

// CadenaBlockchainResponse representa el historial completo de eventos de un contrato
type CadenaBlockchainResponse struct {
	ContractAddress string              `json:"contractAddress"`
	LoteID          string              `json:"loteId"`
	TotalEventos    int                 `json:"totalEventos"`
	Eventos         []EventoBlockchain  `json:"eventos"`
}