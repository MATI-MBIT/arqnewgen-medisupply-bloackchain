package services

import (
	"CrearLoteMicro/models"
	"context"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// ObtenerCadenaBlockchainOptimizada versión optimizada para RPC gratuitos con limitaciones estrictas
func (bs *BlockchainService) ObtenerCadenaBlockchainOptimizada(contractAddress string) (*models.CadenaBlockchainResponse, error) {
	// Parsear ABI
	parsedABI, err := abi.JSON(strings.NewReader(contractABI))
	if err != nil {
		return nil, fmt.Errorf("error parseando ABI: %v", err)
	}

	// Dirección del contrato
	contractAddr := common.HexToAddress(contractAddress)

	// Verificar que el contrato existe
	code, err := bs.Client.CodeAt(context.Background(), contractAddr, nil)
	if err != nil {
		return nil, fmt.Errorf("error verificando contrato: %v", err)
	}
	if len(code) == 0 {
		return nil, fmt.Errorf("no se encontró contrato en la dirección especificada")
	}

	// Obtener información básica del lote desde el contrato
	loteID, err := bs.obtenerLoteIDDelContrato(contractAddr, parsedABI)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo loteID: %v", err)
	}

	// Buscar eventos usando una estrategia conservadora
	eventos, err := bs.buscarEventosConservadora(contractAddr, parsedABI)
	if err != nil {
		return nil, fmt.Errorf("error buscando eventos: %v", err)
	}

	response := &models.CadenaBlockchainResponse{
		ContractAddress: contractAddress,
		LoteID:          loteID,
		TotalEventos:    len(eventos),
		Eventos:         eventos,
	}

	return response, nil
}

// obtenerLoteIDDelContrato obtiene el loteID directamente del contrato
func (bs *BlockchainService) obtenerLoteIDDelContrato(contractAddr common.Address, parsedABI abi.ABI) (string, error) {
	contract := bind.NewBoundContract(contractAddr, parsedABI, bs.Client, bs.Client, bs.Client)
	callOpts := &bind.CallOpts{Context: context.Background()}

	var loteIdResult []interface{}
	err := contract.Call(callOpts, &loteIdResult, "loteId")
	if err != nil {
		return "", err
	}

	if len(loteIdResult) > 0 {
		return loteIdResult[0].(string), nil
	}

	return "", fmt.Errorf("no se pudo obtener loteID")
}

// buscarEventosConservadora busca eventos usando una estrategia muy conservadora para RPC gratuitos
func (bs *BlockchainService) buscarEventosConservadora(contractAddr common.Address, parsedABI abi.ABI) ([]models.EventoBlockchain, error) {
	var eventos []models.EventoBlockchain

	// Obtener bloque actual
	latestBlock, err := bs.Client.BlockNumber(context.Background())
	if err != nil {
		return nil, err
	}

	// Buscar en los últimos 1000 bloques solamente para evitar limitaciones
	startBlock := latestBlock
	if latestBlock > 1000 {
		startBlock = latestBlock - 1000
	} else {
		startBlock = 0
	}

	// Buscar en lotes de 10 bloques (máximo permitido por RPC gratuitos)
	const batchSize = 10

	for fromBlock := startBlock; fromBlock <= latestBlock; fromBlock += batchSize {
		toBlock := fromBlock + batchSize - 1
		if toBlock > latestBlock {
			toBlock = latestBlock
		}

		query := ethereum.FilterQuery{
			FromBlock: big.NewInt(int64(fromBlock)),
			ToBlock:   big.NewInt(int64(toBlock)),
			Addresses: []common.Address{contractAddr},
		}

		logs, err := bs.Client.FilterLogs(context.Background(), query)
		if err != nil {
			// Si falla, continuar con el siguiente lote
			continue
		}

		// Procesar logs encontrados
		for _, vLog := range logs {
			evento, err := bs.procesarLog(vLog, parsedABI)
			if err != nil {
				continue // Continuar con el siguiente log si hay error
			}
			eventos = append(eventos, evento)
		}
	}

	return eventos, nil
}

// procesarLog procesa un log individual y lo convierte en EventoBlockchain
func (bs *BlockchainService) procesarLog(vLog types.Log, parsedABI abi.ABI) (models.EventoBlockchain, error) {
	// Obtener información del bloque para el timestamp
	block, err := bs.Client.BlockByNumber(context.Background(), big.NewInt(int64(vLog.BlockNumber)))
	if err != nil {
		return models.EventoBlockchain{}, err
	}

	evento := models.EventoBlockchain{
		BlockNumber: vLog.BlockNumber,
		TxHash:      vLog.TxHash.Hex(),
		Timestamp:   block.Time(),
		Datos:       make(map[string]interface{}),
	}

	// Procesar según el tipo de evento
	switch vLog.Topics[0] {
	case parsedABI.Events["LoteCreado"].ID:
		evento.TipoEvento = "LoteCreado"
		
		var loteCreado struct {
			LoteId             string
			Fabricante         common.Address
			TemperaturaMinima  int8
			TemperaturaMaxima  int8
		}
		
		err := parsedABI.UnpackIntoInterface(&loteCreado, "LoteCreado", vLog.Data)
		if err == nil {
			evento.Datos["loteId"] = loteCreado.LoteId
			evento.Datos["fabricante"] = loteCreado.Fabricante.Hex()
			evento.Datos["temperaturaMinima"] = loteCreado.TemperaturaMinima
			evento.Datos["temperaturaMaxima"] = loteCreado.TemperaturaMaxima
		}

	case parsedABI.Events["CustodiaTransferida"].ID:
		evento.TipoEvento = "CustodiaTransferida"
		
		if len(vLog.Topics) >= 3 {
			propietarioAnterior := common.HexToAddress(vLog.Topics[1].Hex())
			nuevoPropietario := common.HexToAddress(vLog.Topics[2].Hex())
			
			evento.Datos["propietarioAnterior"] = propietarioAnterior.Hex()
			evento.Datos["nuevoPropietario"] = nuevoPropietario.Hex()
		}

	case parsedABI.Events["LoteComprometido"].ID:
		evento.TipoEvento = "LoteComprometido"
		
		var loteComprometido struct {
			TemperaturaRegistrada int8
			Motivo               string
		}
		
		err := parsedABI.UnpackIntoInterface(&loteComprometido, "LoteComprometido", vLog.Data)
		if err == nil {
			evento.Datos["temperaturaRegistrada"] = loteComprometido.TemperaturaRegistrada
			evento.Datos["motivo"] = loteComprometido.Motivo
		}
	}

	return evento, nil
}