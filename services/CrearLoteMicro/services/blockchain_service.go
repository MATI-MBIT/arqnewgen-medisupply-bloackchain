package services

import (
	"CrearLoteMicro/assets/contracts"
	"CrearLoteMicro/models"
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

type BlockchainService struct {
	Client  *ethclient.Client
	chainID *big.Int
}

// Funciones para obtener ABI y Bytecode desde assets
func getContractABI() string {
	return contracts.GetLoteTracingABI()
}

func getContractBytecode() string {
	return contracts.GetLoteTracingBytecode()
}

func NewBlockchainService(rpcURL string, chainID int64) (*BlockchainService, error) {
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		return nil, fmt.Errorf("error conectando a la blockchain: %v", err)
	}

	return &BlockchainService{
		Client:  client,
		chainID: big.NewInt(chainID),
	}, nil
}

func (bs *BlockchainService) DeployContract(privateKeyHex, loteID string, tempMin, tempMax int8) (string, string, error) {
	// Parsear la clave privada
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return "", "", fmt.Errorf("error parseando clave privada: %v", err)
	}

	// Obtener la dirección pública
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return "", "", fmt.Errorf("error obteniendo clave pública")
	}
	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	// Obtener nonce
	nonce, err := bs.Client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		return "", "", fmt.Errorf("error obteniendo nonce: %v", err)
	}

	// Configurar gas
	gasPrice, err := bs.Client.SuggestGasPrice(context.Background())
	if err != nil {
		return "", "", fmt.Errorf("error obteniendo gas price: %v", err)
	}

	// Parsear ABI
	parsedABI, err := abi.JSON(strings.NewReader(getContractABI()))
	if err != nil {
		return "", "", fmt.Errorf("error parseando ABI: %v", err)
	}

	// Crear transactor
	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, bs.chainID)
	if err != nil {
		return "", "", fmt.Errorf("error creando transactor: %v", err)
	}

	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0)
	auth.GasLimit = uint64(3000000)
	auth.GasPrice = gasPrice

	// Preparar datos del constructor
	input, err := parsedABI.Pack("", loteID, tempMin, tempMax)
	if err != nil {
		return "", "", fmt.Errorf("error empaquetando datos del constructor: %v", err)
	}

	// Crear transacción de deploy
	data := append(common.FromHex(getContractBytecode()), input...)
	
	tx := types.NewContractCreation(nonce, big.NewInt(0), 3000000, gasPrice, data)
	
	// Firmar transacción
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(bs.chainID), privateKey)
	if err != nil {
		return "", "", fmt.Errorf("error firmando transacción: %v", err)
	}

	// Enviar transacción
	err = bs.Client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return "", "", fmt.Errorf("error enviando transacción: %v", err)
	}

	// Calcular dirección del contrato
	contractAddress := crypto.CreateAddress(fromAddress, nonce)
	
	fmt.Printf("[DEBUG] Deploy completado:\n")
	fmt.Printf("[DEBUG] - From Address: %s\n", fromAddress.Hex())
	fmt.Printf("[DEBUG] - Nonce usado: %d\n", nonce)
	fmt.Printf("[DEBUG] - Contract Address calculada: %s\n", contractAddress.Hex())
	fmt.Printf("[DEBUG] - Transaction Hash: %s\n", signedTx.Hash().Hex())
	fmt.Printf("[DEBUG] - Gas Limit: %d\n", signedTx.Gas())
	fmt.Printf("[DEBUG] - Gas Price: %s\n", signedTx.GasPrice().String())

	return contractAddress.Hex(), signedTx.Hash().Hex(), nil
}

func (bs *BlockchainService) RegistrarTemperatura(privateKeyHex, contractAddress string, tempMin, tempMax int8) (string, error) {
	// Parsear la clave privada
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return "", fmt.Errorf("error parseando clave privada: %v", err)
	}

	// Obtener la dirección pública
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return "", fmt.Errorf("error obteniendo clave pública")
	}
	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	// Obtener nonce
	nonce, err := bs.Client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		return "", fmt.Errorf("error obteniendo nonce: %v", err)
	}

	// Configurar gas
	gasPrice, err := bs.Client.SuggestGasPrice(context.Background())
	if err != nil {
		return "", fmt.Errorf("error obteniendo gas price: %v", err)
	}

	// Parsear ABI
	parsedABI, err := abi.JSON(strings.NewReader(getContractABI()))
	if err != nil {
		return "", fmt.Errorf("error parseando ABI: %v", err)
	}

	// Preparar datos de la función
	data, err := parsedABI.Pack("registrarTemperatura", tempMin, tempMax)
	if err != nil {
		return "", fmt.Errorf("error empaquetando datos de la función: %v", err)
	}

	// Crear transacción
	toAddress := common.HexToAddress(contractAddress)
	tx := types.NewTransaction(nonce, toAddress, big.NewInt(0), 300000, gasPrice, data)

	// Firmar transacción
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(bs.chainID), privateKey)
	if err != nil {
		return "", fmt.Errorf("error firmando transacción: %v", err)
	}

	// Enviar transacción
	err = bs.Client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return "", fmt.Errorf("error enviando transacción: %v", err)
	}

	return signedTx.Hash().Hex(), nil
}

func (bs *BlockchainService) TransferirCustodia(privateKeyHex, contractAddress, nuevoPropietario string) (string, error) {
	// Parsear la clave privada
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return "", fmt.Errorf("error parseando clave privada: %v", err)
	}

	// Obtener la dirección pública
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return "", fmt.Errorf("error obteniendo clave pública")
	}
	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	// Obtener nonce
	nonce, err := bs.Client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		return "", fmt.Errorf("error obteniendo nonce: %v", err)
	}

	// Configurar gas
	gasPrice, err := bs.Client.SuggestGasPrice(context.Background())
	if err != nil {
		return "", fmt.Errorf("error obteniendo gas price: %v", err)
	}

	// Parsear ABI
	parsedABI, err := abi.JSON(strings.NewReader(getContractABI()))
	if err != nil {
		return "", fmt.Errorf("error parseando ABI: %v", err)
	}

	// Preparar datos de la función
	nuevoPropietarioAddr := common.HexToAddress(nuevoPropietario)
	data, err := parsedABI.Pack("transferirCustodia", nuevoPropietarioAddr)
	if err != nil {
		return "", fmt.Errorf("error empaquetando datos de la función: %v", err)
	}

	// Crear transacción
	toAddress := common.HexToAddress(contractAddress)
	tx := types.NewTransaction(nonce, toAddress, big.NewInt(0), 300000, gasPrice, data)

	// Firmar transacción
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(bs.chainID), privateKey)
	if err != nil {
		return "", fmt.Errorf("error firmando transacción: %v", err)
	}

	// Enviar transacción
	err = bs.Client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return "", fmt.Errorf("error enviando transacción: %v", err)
	}

	return signedTx.Hash().Hex(), nil
}

func (bs *BlockchainService) ObtenerInfoLote(contractAddress string) (*models.LoteInfoResponse, error) {
	fmt.Printf("[DEBUG] Iniciando ObtenerInfoLote para dirección: %s\n", contractAddress)
	
	// Parsear ABI
	parsedABI, err := abi.JSON(strings.NewReader(getContractABI()))
	if err != nil {
		fmt.Printf("[ERROR] Error parseando ABI: %v\n", err)
		return nil, fmt.Errorf("error parseando ABI: %v", err)
	}
	fmt.Printf("[DEBUG] ABI parseado exitosamente\n")

	// Dirección del contrato
	contractAddr := common.HexToAddress(contractAddress)
	fmt.Printf("[DEBUG] Dirección del contrato convertida: %s\n", contractAddr.Hex())

	// Verificar que el contrato existe y tiene código
	code, err := bs.Client.CodeAt(context.Background(), contractAddr, nil)
	if err != nil {
		fmt.Printf("[ERROR] Error obteniendo código del contrato: %v\n", err)
		return nil, fmt.Errorf("error verificando contrato: %v", err)
	}
	fmt.Printf("[DEBUG] Código del contrato obtenido, longitud: %d bytes\n", len(code))
	
	if len(code) == 0 {
		fmt.Printf("[ERROR] No se encontró código en la dirección del contrato\n")
		return nil, fmt.Errorf("no se encontró contrato en la dirección especificada")
	}

	// Obtener el número de bloque actual para contexto
	currentBlock, err := bs.Client.BlockNumber(context.Background())
	if err != nil {
		fmt.Printf("[WARNING] No se pudo obtener el bloque actual: %v\n", err)
	} else {
		fmt.Printf("[DEBUG] Bloque actual: %d\n", currentBlock)
	}

	// Crear instancia del contrato para llamadas de solo lectura
	contract := bind.NewBoundContract(contractAddr, parsedABI, bs.Client, bs.Client, bs.Client)
	fmt.Printf("[DEBUG] Instancia del contrato creada\n")

	// Realizar llamadas a las funciones públicas del contrato
	callOpts := &bind.CallOpts{Context: context.Background()}
	fmt.Printf("[DEBUG] CallOpts configurado, iniciando llamadas a funciones\n")

	// Obtener loteId
	fmt.Printf("[DEBUG] Intentando obtener loteId...\n")
	var loteIdResult []interface{}
	err = contract.Call(callOpts, &loteIdResult, "loteId")
	if err != nil {
		fmt.Printf("[ERROR] Error obteniendo loteId: %v\n", err)
		return nil, fmt.Errorf("error obteniendo loteId: %v", err)
	}
	fmt.Printf("[DEBUG] loteIdResult obtenido: %+v\n", loteIdResult)
	loteId := loteIdResult[0].(string)
	fmt.Printf("[DEBUG] loteId extraído: %s\n", loteId)

	// Obtener fabricante
	fmt.Printf("[DEBUG] Intentando obtener fabricante...\n")
	var fabricanteResult []interface{}
	err = contract.Call(callOpts, &fabricanteResult, "fabricante")
	if err != nil {
		fmt.Printf("[ERROR] Error obteniendo fabricante: %v\n", err)
		return nil, fmt.Errorf("error obteniendo fabricante: %v", err)
	}
	fmt.Printf("[DEBUG] fabricanteResult obtenido: %+v\n", fabricanteResult)
	fabricante := fabricanteResult[0].(common.Address)
	fmt.Printf("[DEBUG] fabricante extraído: %s\n", fabricante.Hex())

	// Obtener propietario actual
	var propietarioResult []interface{}
	err = contract.Call(callOpts, &propietarioResult, "propietarioActual")
	if err != nil {
		return nil, fmt.Errorf("error obteniendo propietario actual: %v", err)
	}
	propietarioActual := propietarioResult[0].(common.Address)

	// Obtener temperatura mínima
	var tempMinResult []interface{}
	err = contract.Call(callOpts, &tempMinResult, "temperaturaMinima")
	if err != nil {
		return nil, fmt.Errorf("error obteniendo temperatura mínima: %v", err)
	}
	temperaturaMinima := tempMinResult[0].(int8)

	// Obtener temperatura máxima
	var tempMaxResult []interface{}
	err = contract.Call(callOpts, &tempMaxResult, "temperaturaMaxima")
	if err != nil {
		return nil, fmt.Errorf("error obteniendo temperatura máxima: %v", err)
	}
	temperaturaMaxima := tempMaxResult[0].(int8)

	// Obtener estado comprometido
	var comprometidoResult []interface{}
	err = contract.Call(callOpts, &comprometidoResult, "comprometido")
	if err != nil {
		return nil, fmt.Errorf("error obteniendo estado comprometido: %v", err)
	}
	comprometido := comprometidoResult[0].(bool)

	// Obtener temperatura registrada mínima
	var tempRegMinResult []interface{}
	err = contract.Call(callOpts, &tempRegMinResult, "tempRegMinima")
	if err != nil {
		return nil, fmt.Errorf("error obteniendo temperatura registrada mínima: %v", err)
	}
	tempRegMinima := tempRegMinResult[0].(int8)

	// Obtener temperatura registrada máxima
	var tempRegMaxResult []interface{}
	err = contract.Call(callOpts, &tempRegMaxResult, "tempRegMaxima")
	if err != nil {
		return nil, fmt.Errorf("error obteniendo temperatura registrada máxima: %v", err)
	}
	tempRegMaxima := tempRegMaxResult[0].(int8)

	// Crear respuesta
	response := &models.LoteInfoResponse{
		LoteID:            loteId,
		Fabricante:        fabricante.Hex(),
		PropietarioActual: propietarioActual.Hex(),
		TemperaturaMinima: temperaturaMinima,
		TemperaturaMaxima: temperaturaMaxima,
		TempRegMinima:     tempRegMinima,
		TempRegMaxima:     tempRegMaxima,
		Comprometido:      comprometido,
		ContractAddress:   contractAddress,
	}

	return response, nil
}

func (bs *BlockchainService) ObtenerCadenaBlockchain(contractAddress string) (*models.CadenaBlockchainResponse, error) {
	// Parsear ABI
	parsedABI, err := abi.JSON(strings.NewReader(getContractABI()))
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

	// Obtener el bloque actual
	latestBlock, err := bs.Client.BlockNumber(context.Background())
	if err != nil {
		return nil, fmt.Errorf("error obteniendo bloque actual: %v", err)
	}

	// Obtener todos los logs usando consultas por lotes
	logs, err := bs.obtenerLogsPorLotes(contractAddr, latestBlock)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo logs: %v", err)
	}

	// Procesar eventos
	var eventos []models.EventoBlockchain
	var loteID string

	for _, vLog := range logs {
		// Obtener información del bloque para el timestamp
		block, err := bs.Client.BlockByNumber(context.Background(), big.NewInt(int64(vLog.BlockNumber)))
		if err != nil {
			continue // Continuar con el siguiente log si hay error
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
			
			// Decodificar evento LoteCreado
			var loteCreado struct {
				LoteId             string
				Fabricante         common.Address
				TemperaturaMinima  int8
				TemperaturaMaxima  int8
			}
			
			err := parsedABI.UnpackIntoInterface(&loteCreado, "LoteCreado", vLog.Data)
			if err == nil {
				loteID = loteCreado.LoteId // Guardar el loteID para la respuesta
				evento.Datos["loteId"] = loteCreado.LoteId
				evento.Datos["fabricante"] = loteCreado.Fabricante.Hex()
				evento.Datos["temperaturaMinima"] = loteCreado.TemperaturaMinima
				evento.Datos["temperaturaMaxima"] = loteCreado.TemperaturaMaxima
			}

		case parsedABI.Events["CustodiaTransferida"].ID:
			evento.TipoEvento = "CustodiaTransferida"
			
			// Los topics indexados están en vLog.Topics[1] y vLog.Topics[2]
			if len(vLog.Topics) >= 3 {
				propietarioAnterior := common.HexToAddress(vLog.Topics[1].Hex())
				nuevoPropietario := common.HexToAddress(vLog.Topics[2].Hex())
				
				evento.Datos["propietarioAnterior"] = propietarioAnterior.Hex()
				evento.Datos["nuevoPropietario"] = nuevoPropietario.Hex()
			}
			
			// Decodificar el campo comprometido del evento
			var custodiaTransferida struct {
				Comprometido bool
			}
			
			err := parsedABI.UnpackIntoInterface(&custodiaTransferida, "CustodiaTransferida", vLog.Data)
			if err == nil {
				evento.Datos["comprometido"] = custodiaTransferida.Comprometido
			}

		case parsedABI.Events["LoteComprometido"].ID:
			evento.TipoEvento = "LoteComprometido"
			
			// El propietario está indexado en vLog.Topics[1]
			if len(vLog.Topics) >= 2 {
				propietario := common.HexToAddress(vLog.Topics[1].Hex())
				evento.Datos["propietario"] = propietario.Hex()
			}
			
			// Decodificar evento LoteComprometido actualizado
			var loteComprometido struct {
				TempMin      int8
				TempMax      int8
				Comprometido bool
				Motivo       string
			}
			
			err := parsedABI.UnpackIntoInterface(&loteComprometido, "LoteComprometido", vLog.Data)
			if err == nil {
				evento.Datos["tempMin"] = loteComprometido.TempMin
				evento.Datos["tempMax"] = loteComprometido.TempMax
				evento.Datos["comprometido"] = loteComprometido.Comprometido
				evento.Datos["motivo"] = loteComprometido.Motivo
			}
		}

		eventos = append(eventos, evento)
	}

	// Si no encontramos el loteID en los eventos, intentar obtenerlo del contrato
	if loteID == "" {
		// Crear instancia del contrato para llamadas de solo lectura
		contract := bind.NewBoundContract(contractAddr, parsedABI, bs.Client, bs.Client, bs.Client)
		callOpts := &bind.CallOpts{Context: context.Background()}

		var loteIdResult []interface{}
		err = contract.Call(callOpts, &loteIdResult, "loteId")
		if err == nil && len(loteIdResult) > 0 {
			loteID = loteIdResult[0].(string)
		}
	}

	response := &models.CadenaBlockchainResponse{
		ContractAddress: contractAddress,
		LoteID:          loteID,
		TotalEventos:    len(eventos),
		Eventos:         eventos,
	}

	return response, nil
}

// obtenerLogsPorLotes obtiene logs del contrato usando consultas por lotes para evitar limitaciones de RPC gratuitos
func (bs *BlockchainService) obtenerLogsPorLotes(contractAddr common.Address, latestBlock uint64) ([]types.Log, error) {
	// Estrategia optimizada: buscar desde los bloques más recientes hacia atrás
	// ya que los contratos suelen ser recientes
	const batchSize = 10 // Usar lotes pequeños para compatibilidad con RPC gratuitos
	var allLogs []types.Log
	
	// Empezar desde los últimos 50,000 bloques para optimizar la búsqueda
	startBlock := uint64(0)
	if latestBlock > 50000 {
		startBlock = latestBlock - 50000
	}

	// Buscar en lotes pequeños
	for fromBlock := startBlock; fromBlock <= latestBlock; fromBlock += batchSize {
		toBlock := fromBlock + batchSize - 1
		if toBlock > latestBlock {
			toBlock = latestBlock
		}

		// Crear filtro para este lote
		query := ethereum.FilterQuery{
			FromBlock: big.NewInt(int64(fromBlock)),
			ToBlock:   big.NewInt(int64(toBlock)),
			Addresses: []common.Address{contractAddr},
		}

		// Obtener logs para este lote
		logs, err := bs.Client.FilterLogs(context.Background(), query)
		if err != nil {
			// Si aún falla, intentar bloque por bloque
			singleBlockLogs, singleErr := bs.obtenerLogsBloqueIndividual(contractAddr, fromBlock, toBlock)
			if singleErr != nil {
				return nil, fmt.Errorf("error en lote %d-%d: %v", fromBlock, toBlock, err)
			}
			allLogs = append(allLogs, singleBlockLogs...)
		} else {
			allLogs = append(allLogs, logs...)
		}
	}

	return allLogs, nil
}

// obtenerLogsBloqueIndividual maneja casos extremos donde necesitamos consultar bloque por bloque
func (bs *BlockchainService) obtenerLogsBloqueIndividual(contractAddr common.Address, fromBlock, toBlock uint64) ([]types.Log, error) {
	var allLogs []types.Log

	for block := fromBlock; block <= toBlock; block++ {
		query := ethereum.FilterQuery{
			FromBlock: big.NewInt(int64(block)),
			ToBlock:   big.NewInt(int64(block)),
			Addresses: []common.Address{contractAddr},
		}

		logs, err := bs.Client.FilterLogs(context.Background(), query)
		if err != nil {
			// Si falla incluso con un solo bloque, continuar con el siguiente
			continue
		}

		allLogs = append(allLogs, logs...)
	}

	return allLogs, nil
}

// DiagnosticarContrato proporciona información detallada sobre el estado de un contrato
func (bs *BlockchainService) DiagnosticarContrato(contractAddress string) (map[string]interface{}, error) {
	diagnostico := make(map[string]interface{})
	
	fmt.Printf("[DEBUG] Iniciando diagnóstico para contrato: %s\n", contractAddress)
	
	// Dirección del contrato
	contractAddr := common.HexToAddress(contractAddress)
	diagnostico["contractAddress"] = contractAddr.Hex()
	
	// Obtener bloque actual
	currentBlock, err := bs.Client.BlockNumber(context.Background())
	if err != nil {
		diagnostico["currentBlockError"] = err.Error()
	} else {
		diagnostico["currentBlock"] = currentBlock
	}
	
	// Verificar código del contrato
	code, err := bs.Client.CodeAt(context.Background(), contractAddr, nil)
	if err != nil {
		diagnostico["codeError"] = err.Error()
	} else {
		diagnostico["codeLength"] = len(code)
		diagnostico["hasCode"] = len(code) > 0
		if len(code) > 0 {
			diagnostico["codeHash"] = fmt.Sprintf("0x%x", crypto.Keccak256(code))
		}
	}
	
	// Verificar balance del contrato
	balance, err := bs.Client.BalanceAt(context.Background(), contractAddr, nil)
	if err != nil {
		diagnostico["balanceError"] = err.Error()
	} else {
		diagnostico["balance"] = balance.String()
	}
	
	// Verificar nonce del contrato
	nonce, err := bs.Client.NonceAt(context.Background(), contractAddr, nil)
	if err != nil {
		diagnostico["nonceError"] = err.Error()
	} else {
		diagnostico["nonce"] = nonce
	}
	
	// Si el contrato tiene código, intentar llamadas básicas
	if len(code) > 0 {
		diagnostico["contractCalls"] = bs.probarLlamadasContrato(contractAddr)
	}
	
	// Buscar transacciones recientes relacionadas con esta dirección
	diagnostico["recentActivity"] = bs.buscarActividadReciente(contractAddr)
	
	return diagnostico, nil
}

// probarLlamadasContrato intenta hacer llamadas básicas al contrato
func (bs *BlockchainService) probarLlamadasContrato(contractAddr common.Address) map[string]interface{} {
	calls := make(map[string]interface{})
	
	// Parsear ABI
	parsedABI, err := abi.JSON(strings.NewReader(getContractABI()))
	if err != nil {
		calls["abiError"] = err.Error()
		return calls
	}
	
	// Crear instancia del contrato
	contract := bind.NewBoundContract(contractAddr, parsedABI, bs.Client, bs.Client, bs.Client)
	callOpts := &bind.CallOpts{Context: context.Background()}
	
	// Intentar llamar a cada función de solo lectura
	functions := []string{"loteId", "fabricante", "propietarioActual", "temperaturaMinima", "temperaturaMaxima", "tempRegMinima", "tempRegMaxima", "comprometido"}
	
	for _, funcName := range functions {
		var result []interface{}
		err := contract.Call(callOpts, &result, funcName)
		if err != nil {
			calls[funcName+"_error"] = err.Error()
		} else {
			calls[funcName+"_success"] = true
			if len(result) > 0 {
				calls[funcName+"_value"] = fmt.Sprintf("%v", result[0])
			}
		}
	}
	
	return calls
}

// buscarActividadReciente busca transacciones recientes relacionadas con el contrato
func (bs *BlockchainService) buscarActividadReciente(contractAddr common.Address) map[string]interface{} {
	activity := make(map[string]interface{})
	
	// Obtener bloque actual
	currentBlock, err := bs.Client.BlockNumber(context.Background())
	if err != nil {
		activity["error"] = "No se pudo obtener bloque actual: " + err.Error()
		return activity
	}
	
	// Buscar en los últimos 10 bloques solamente (límite de RPC gratuito)
	startBlock := currentBlock
	if currentBlock > 10 {
		startBlock = currentBlock - 10
	} else {
		startBlock = 0
	}
	
	// Buscar logs relacionados con este contrato
	query := ethereum.FilterQuery{
		FromBlock: big.NewInt(int64(startBlock)),
		ToBlock:   big.NewInt(int64(currentBlock)),
		Addresses: []common.Address{contractAddr},
	}
	
	logs, err := bs.Client.FilterLogs(context.Background(), query)
	if err != nil {
		// Si falla, intentar solo el bloque actual
		query.FromBlock = big.NewInt(int64(currentBlock))
		logs, err = bs.Client.FilterLogs(context.Background(), query)
		if err != nil {
			activity["logsError"] = "RPC limitado - no se pueden buscar logs: " + err.Error()
			activity["note"] = "Esto es normal con RPC gratuitos. El contrato puede funcionar correctamente."
		} else {
			activity["recentLogs"] = len(logs)
			activity["searchRange"] = "Solo bloque actual debido a limitaciones RPC"
		}
	} else {
		activity["recentLogs"] = len(logs)
		activity["searchRange"] = fmt.Sprintf("Últimos %d bloques", currentBlock-startBlock+1)
		if len(logs) > 0 {
			activity["latestLogBlock"] = logs[len(logs)-1].BlockNumber
			activity["latestLogTxHash"] = logs[len(logs)-1].TxHash.Hex()
		}
	}
	
	return activity
}