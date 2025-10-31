package handlers

import (
	"CrearLoteMicro/models"
	"CrearLoteMicro/services"
	"CrearLoteMicro/utils"
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type LoteHandler struct {
	blockchainService          *services.BlockchainService
	blockchainWebsocketService *services.BlockchainWebsocketService
}

func NewLoteHandler(blockchainService *services.BlockchainService, blockchainWebsocketService *services.BlockchainWebsocketService) *LoteHandler {
	return &LoteHandler{
		blockchainService:          blockchainService,
		blockchainWebsocketService: blockchainWebsocketService,
	}
}

// CrearLote maneja la creación de un nuevo lote (deploy del contrato)
func (h *LoteHandler) CrearLote(c *gin.Context) {
	var req models.CrearLoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.Response{
			Success: false,
			Message: "Datos de entrada inválidos: " + err.Error(),
		})
		return
	}

	// Validar que la clave privada no esté vacía
	if req.PrivateKey == "" {
		c.JSON(http.StatusBadRequest, models.Response{
			Success: false,
			Message: "La clave privada es requerida",
		})
		return
	}

	// Desplegar el contrato
	contractAddress, txHash, err := h.blockchainService.DeployContract(
		req.PrivateKey,
		req.LoteID,
		req.TemperaturaMin,
		req.TemperaturaMax,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Success: false,
			Message: "Error desplegando contrato: " + err.Error(),
		})
		return
	}
	fmt.Println("deploying contract at:", contractAddress)
	go h.blockchainWebsocketService.StartBlockchainWebsocket(contractAddress)

	response := models.ContractDeployResponse{
		ContractAddress: contractAddress,
		TxHash:          txHash,
		LoteID:          req.LoteID,
	}

	c.JSON(http.StatusOK, models.Response{
		Success: true,
		Message: "Lote creado exitosamente con socket",
		Data:    response,
		TxHash:  txHash,
	})
}

// RegistrarTemperatura maneja el registro de temperatura en un lote existente
func (h *LoteHandler) RegistrarTemperatura(c *gin.Context) {
	var req models.RegistrarTemperaturaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.Response{
			Success: false,
			Message: "Datos de entrada inválidos: " + err.Error(),
		})
		return
	}

	// Validar que la clave privada no esté vacía
	if req.PrivateKey == "" {
		c.JSON(http.StatusBadRequest, models.Response{
			Success: false,
			Message: "La clave privada es requerida",
		})
		return
	}

	// Registrar temperatura
	txHash, err := h.blockchainService.RegistrarTemperatura(
		req.PrivateKey,
		req.ContractAddress,
		req.TempMin,
		req.TempMax,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Success: false,
			Message: "Error registrando temperatura: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Success: true,
		Message: "Temperatura registrada exitosamente",
		TxHash:  txHash,
	})
}

// TransferirCustodia maneja la transferencia de custodia de un lote
func (h *LoteHandler) TransferirCustodia(c *gin.Context) {
	var req models.TransferirCustodiaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.Response{
			Success: false,
			Message: "Datos de entrada inválidos: " + err.Error(),
		})
		return
	}

	// Validar que la clave privada no esté vacía
	if req.PrivateKey == "" {
		c.JSON(http.StatusBadRequest, models.Response{
			Success: false,
			Message: "La clave privada es requerida",
		})
		return
	}

	// Transferir custodia
	txHash, err := h.blockchainService.TransferirCustodia(
		req.PrivateKey,
		req.ContractAddress,
		req.NuevoPropietario,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Success: false,
			Message: "Error transfiriendo custodia: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Success: true,
		Message: "Custodia transferida exitosamente",
		TxHash:  txHash,
	})
}

// CrearNuevoLote maneja la creación de un nuevo lote en un contrato existente
func (h *LoteHandler) CrearNuevoLote(c *gin.Context) {
	var req models.CrearNuevoLoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.Response{
			Success: false,
			Message: "Datos de entrada inválidos: " + err.Error(),
		})
		return
	}

	// Validar que la clave privada no esté vacía
	if req.PrivateKey == "" {
		c.JSON(http.StatusBadRequest, models.Response{
			Success: false,
			Message: "La clave privada es requerida",
		})
		return
	}

	// Validar formato de dirección del contrato
	if len(req.ContractAddress) != 42 || req.ContractAddress[:2] != "0x" {
		c.JSON(http.StatusBadRequest, models.Response{
			Success: false,
			Message: "Formato de dirección de contrato inválido",
		})
		return
	}

	// Crear nuevo lote en el contrato existente
	txHash, err := h.blockchainService.CrearNuevoLote(
		req.PrivateKey,
		req.ContractAddress,
		req.LoteID,
		req.TemperaturaMin,
		req.TemperaturaMax,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Success: false,
			Message: "Error creando nuevo lote: " + err.Error(),
		})
		return
	}

	response := map[string]interface{}{
		"contractAddress": req.ContractAddress,
		"loteId":          req.LoteID,
		"txHash":          txHash,
	}

	c.JSON(http.StatusOK, models.Response{
		Success: true,
		Message: "Nuevo lote creado exitosamente en contrato existente",
		Data:    response,
		TxHash:  txHash,
	})
}

// HealthCheck endpoint para verificar el estado del servicio
func (h *LoteHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, models.Response{
		Success: true,
		Message: "CrearLoteMicro está funcionando correctamente",
	})
}

// ObtenerLote maneja la consulta de información de un lote existente
func (h *LoteHandler) ObtenerLote(c *gin.Context) {
	// Obtener la dirección del contrato desde los parámetros de la URL
	contractAddress := c.Param("contractAddress")

	if contractAddress == "" {
		c.JSON(http.StatusBadRequest, models.Response{
			Success: false,
			Message: "La dirección del contrato es requerida",
		})
		return
	}

	// Validar formato de dirección Ethereum
	if len(contractAddress) != 42 || contractAddress[:2] != "0x" {
		c.JSON(http.StatusBadRequest, models.Response{
			Success: false,
			Message: "Formato de dirección de contrato inválido",
		})
		return
	}

	// Obtener información del lote
	loteInfo, err := h.blockchainService.ObtenerInfoLote(contractAddress)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Success: false,
			Message: "Error obteniendo información del lote: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Success: true,
		Message: "Información del lote obtenida exitosamente",
		Data:    loteInfo,
	})
}

// VerificarConexion endpoint para verificar la conexión a Sepolia
func (h *LoteHandler) VerificarConexion(c *gin.Context) {
	// Obtener el número de bloque actual para verificar conexión
	blockNumber, err := h.blockchainService.Client.BlockNumber(context.Background())
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Success: false,
			Message: "Error conectando a Sepolia: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Success: true,
		Message: "Conexión a Sepolia exitosa",
		Data: map[string]interface{}{
			"blockNumber": blockNumber,
			"chainId":     "11155111", // Sepolia Chain ID
		},
	})
}

// ObtenerCadenaBlockchain maneja la consulta del historial completo de eventos de un contrato
func (h *LoteHandler) ObtenerCadenaBlockchain(c *gin.Context) {
	// Obtener la dirección del contrato desde los parámetros de la URL
	contractAddress := c.Param("contractAddress")

	if contractAddress == "" {
		c.JSON(http.StatusBadRequest, models.Response{
			Success: false,
			Message: "La dirección del contrato es requerida",
		})
		return
	}

	// Validar formato de dirección Ethereum
	if len(contractAddress) != 42 || contractAddress[:2] != "0x" {
		c.JSON(http.StatusBadRequest, models.Response{
			Success: false,
			Message: "Formato de dirección de contrato inválido",
		})
		return
	}

	// Obtener la cadena completa de eventos del contrato usando la versión optimizada
	cadenaBlockchain, err := h.blockchainService.ObtenerCadenaBlockchainOptimizada(contractAddress)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Success: false,
			Message: "Error obteniendo cadena blockchain: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Success: true,
		Message: "Cadena blockchain obtenida exitosamente",
		Data:    cadenaBlockchain,
	})
}

// DecodificarInputData maneja la decodificación de input data de transacciones
func (h *LoteHandler) DecodificarInputData(c *gin.Context) {
	// Obtener el input data desde el cuerpo de la request o parámetro
	var request struct {
		InputData string `json:"inputData" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, models.Response{
			Success: false,
			Message: "Input data es requerido: " + err.Error(),
		})
		return
	}

	// Decodificar el input data
	decoded, err := utils.DecodeInputData(request.InputData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Success: false,
			Message: "Error decodificando input data: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Success: true,
		Message: "Input data decodificado exitosamente",
		Data:    decoded,
	})
}

// DecodificarInputDataEspecifico decodifica el input data específico que mencionaste
func (h *LoteHandler) DecodificarInputDataEspecifico(c *gin.Context) {
	// Decodificar el input data específico
	decoded, err := utils.DecodeSpecificInputData()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Success: false,
			Message: "Error decodificando input data: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Success: true,
		Message: "Input data específico decodificado exitosamente",
		Data:    decoded,
	})
}

// ObtenerSignaturesFunciones retorna todas las signatures de funciones del contrato
func (h *LoteHandler) ObtenerSignaturesFunciones(c *gin.Context) {
	signatures := utils.GetFunctionSignatures()

	c.JSON(http.StatusOK, models.Response{
		Success: true,
		Message: "Signatures de funciones obtenidas exitosamente",
		Data:    signatures,
	})
}

// DiagnosticarContrato proporciona información detallada sobre el estado de un contrato
func (h *LoteHandler) DiagnosticarContrato(c *gin.Context) {
	// Obtener la dirección del contrato desde los parámetros de la URL
	contractAddress := c.Param("contractAddress")

	if contractAddress == "" {
		c.JSON(http.StatusBadRequest, models.Response{
			Success: false,
			Message: "La dirección del contrato es requerida",
		})
		return
	}

	// Validar formato de dirección Ethereum
	if len(contractAddress) != 42 || contractAddress[:2] != "0x" {
		c.JSON(http.StatusBadRequest, models.Response{
			Success: false,
			Message: "Formato de dirección de contrato inválido",
		})
		return
	}

	// Realizar diagnóstico detallado
	diagnostico, err := h.blockchainService.DiagnosticarContrato(contractAddress)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Success: false,
			Message: "Error diagnosticando contrato: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Success: true,
		Message: "Diagnóstico del contrato completado",
		Data:    diagnostico,
	})
}
