package services

import (
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

// ABI del contrato LoteTracing - Compilado con Hardhat
const contractABI = `[
	{
		"inputs": [
			{"internalType": "string", "name": "_loteId", "type": "string"},
			{"internalType": "int8", "name": "_tempMin", "type": "int8"},
			{"internalType": "int8", "name": "_tempMax", "type": "int8"}
		],
		"stateMutability": "nonpayable",
		"type": "constructor"
	},
	{
		"anonymous": false,
		"inputs": [
			{"indexed": true, "internalType": "address", "name": "propietarioAnterior", "type": "address"},
			{"indexed": true, "internalType": "address", "name": "nuevoPropietario", "type": "address"}
		],
		"name": "CustodiaTransferida",
		"type": "event"
	},
	{
		"anonymous": false,
		"inputs": [
			{"indexed": false, "internalType": "int8", "name": "temperaturaRegistrada", "type": "int8"},
			{"indexed": false, "internalType": "string", "name": "motivo", "type": "string"}
		],
		"name": "LoteComprometido",
		"type": "event"
	},
	{
		"anonymous": false,
		"inputs": [
			{"indexed": true, "internalType": "string", "name": "loteId", "type": "string"},
			{"indexed": true, "internalType": "address", "name": "fabricante", "type": "address"},
			{"indexed": false, "internalType": "int8", "name": "temperaturaMinima", "type": "int8"},
			{"indexed": false, "internalType": "int8", "name": "temperaturaMaxima", "type": "int8"}
		],
		"name": "LoteCreado",
		"type": "event"
	},
	{
		"inputs": [],
		"name": "comprometido",
		"outputs": [{"internalType": "bool", "name": "", "type": "bool"}],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [],
		"name": "fabricante",
		"outputs": [{"internalType": "address", "name": "", "type": "address"}],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [],
		"name": "loteId",
		"outputs": [{"internalType": "string", "name": "", "type": "string"}],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [],
		"name": "propietarioActual",
		"outputs": [{"internalType": "address", "name": "", "type": "address"}],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [
			{"internalType": "int8", "name": "_temperatura", "type": "int8"}
		],
		"name": "registrarTemperatura",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [],
		"name": "temperaturaMaxima",
		"outputs": [{"internalType": "int8", "name": "", "type": "int8"}],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [],
		"name": "temperaturaMinima",
		"outputs": [{"internalType": "int8", "name": "", "type": "int8"}],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [
			{"internalType": "address", "name": "_nuevoPropietario", "type": "address"}
		],
		"name": "transferirCustodia",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	}
]`

// Bytecode del contrato (necesario para el deploy) - Compilado con Hardhat
const contractBytecode = "0x60e060405234801561000f575f5ffd5b506040516108e73803806108e783398101604081905261002e916100e9565b5f610039848261023d565b50336080819052600180545f85810b60a05284900b60c0526001600160a81b03191660ff60a01b1983161790556040516100749085906102f7565b604080519182900382205f86810b845285900b6020840152917f614b7bdb598a394ea5f748900a1d99d9db2adc2a0662e25111fbf0b6d06dbf74910160405180910390a350505061030d565b634e487b7160e01b5f52604160045260245ffd5b80515f81900b81146100e4575f5ffd5b919050565b5f5f5f606084860312156100fb575f5ffd5b83516001600160401b03811115610110575f5ffd5b8401601f81018613610120575f5ffd5b80516001600160401b03811115610139576101396100c0565b604051601f8201601f19908116603f011681016001600160401b0381118282101715610167576101676100c0565b60405281815282820160200188101561017e575f5ffd5b8160208401602083015e5f602083830101528095505050506101a2602085016100d4565b91506101b0604085016100d4565b90509250925092565b600181811c908216806101cd57607f821691505b6020821081036101eb57634e487b7160e01b5f52602260045260245ffd5b50919050565b601f82111561023857805f5260205f20601f840160051c810160208510156102165750805b601f840160051c820191505b81811015610235575f8155600101610222565b50505b505050565b81516001600160401b03811115610256576102566100c0565b61026a8161026484546101b9565b846101f1565b6020601f82116001811461029c575f83156102855750848201515b5f19600385901b1c1916600184901b178455610235565b5f84815260208120601f198516915b828110156102cb57878501518255602094850194600190920191016102ab565b50848210156102e857868401515f19600387901b60f8161c191681555b50505050600190811b01905550565b5f82518060208501845e5f920191825250919050565b60805160a05160c0516105a46103435f395f818160a301526103ae01525f8181610157015261038101525f60e101526105a45ff3fe608060405234801561000f575f5ffd5b5060043610610085575f3560e01c806395defb561161005857806395defb561461013f578063af1e625314610152578063d48cf49014610179578063f7b5b4e91461018e575f5ffd5b80631ccbe36b146100895780632ba6b7521461009e57806346ed76f1146100dc57806386b7d1e01461011b575b5f5ffd5b61009c610097366004610465565b6101a1565b005b6100c57f000000000000000000000000000000000000000000000000000000000000000081565b6040515f9190910b81526020015b60405180910390f35b6101037f000000000000000000000000000000000000000000000000000000000000000081565b6040516001600160a01b0390911681526020016100d3565b60015461012f90600160a01b900460ff1681565b60405190151581526020016100d3565b600154610103906001600160a01b031681565b6100c57f000000000000000000000000000000000000000000000000000000000000000081565b610181610270565b6040516100d39190610492565b61009c61019c3660046104c7565b6102fb565b6001546001600160a01b031633146101d45760405162461bcd60e51b81526004016101cb906104e6565b60405180910390fd5b6001600160a01b03811661021f5760405162461bcd60e51b8152602060048201526012602482015271446972656363696f6e20696e76616c69646160701b60448201526064016101cb565b600180546001600160a01b038381166001600160a01b0319831681179093556040519116919082907fc43c743146f2bcef6ee8436a5347b040074a9908c45eadadaca787f0c71990df905f90a35050565b5f805461027c90610536565b80601f01602080910402602001604051908101604052809291908181526020018280546102a890610536565b80156102f35780601f106102ca576101008083540402835291602001916102f3565b820191905f5260205f20905b8154815290600101906020018083116102d657829003601f168201915b505050505081565b6001546001600160a01b031633146103255760405162461bcd60e51b81526004016101cb906104e6565b600154600160a01b900460ff161561037f5760405162461bcd60e51b815260206004820152601c60248201527f456c206c6f7465207961206573746120636f6d70726f6d657469646f0000000060448201526064016101cb565b7f00000000000000000000000000000000000000000000000000000000000000005f0b815f0b12806103d457507f00000000000000000000000000000000000000000000000000000000000000005f0b815f0b135b15610462576001805460ff60a01b1916600160a01b1790556040517f5f157c55726f3dbbb1b632d87789b26afef56340a48a97d6cf2cca5daf352647906104599083905f9190910b8152604060208201819052601a908201527f54656d70657261747572612066756572612064652072616e676f000000000000606082015260800190565b60405180910390a15b50565b5f60208284031215610475575f5ffd5b81356001600160a01b038116811461048b575f5ffd5b9392505050565b602081525f82518060208401528060208501604085015e5f604082850101526040601f19601f83011684010191505092915050565b5f602082840312156104d7575f5ffd5b8135805f0b811461048b575f5ffd5b60208082526030908201527f416363696f6e20736f6c6f207065726d6974696461207061726120656c20707260408201526f1bdc1a595d185c9a5bc81858dd1d585b60821b606082015260800190565b600181811c9082168061054a57607f821691505b60208210810361056857634e487b7160e01b5f52602260045260245ffd5b5091905056fea26469706673582212205a5e88911518703f091238e4c9b360ca923d649b68acce63de7535527607825f64736f6c634300081c0033"

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
	parsedABI, err := abi.JSON(strings.NewReader(contractABI))
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
	data := append(common.FromHex(contractBytecode), input...)
	
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

func (bs *BlockchainService) RegistrarTemperatura(privateKeyHex, contractAddress string, temperatura int8) (string, error) {
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
	parsedABI, err := abi.JSON(strings.NewReader(contractABI))
	if err != nil {
		return "", fmt.Errorf("error parseando ABI: %v", err)
	}

	// Preparar datos de la función
	data, err := parsedABI.Pack("registrarTemperatura", temperatura)
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
	parsedABI, err := abi.JSON(strings.NewReader(contractABI))
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
	parsedABI, err := abi.JSON(strings.NewReader(contractABI))
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

	// Crear respuesta
	response := &models.LoteInfoResponse{
		LoteID:            loteId,
		Fabricante:        fabricante.Hex(),
		PropietarioActual: propietarioActual.Hex(),
		TemperaturaMinima: temperaturaMinima,
		TemperaturaMaxima: temperaturaMaxima,
		Comprometido:      comprometido,
		ContractAddress:   contractAddress,
	}

	return response, nil
}

func (bs *BlockchainService) ObtenerCadenaBlockchain(contractAddress string) (*models.CadenaBlockchainResponse, error) {
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

		case parsedABI.Events["LoteComprometido"].ID:
			evento.TipoEvento = "LoteComprometido"
			
			// Decodificar evento LoteComprometido
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
	parsedABI, err := abi.JSON(strings.NewReader(contractABI))
	if err != nil {
		calls["abiError"] = err.Error()
		return calls
	}
	
	// Crear instancia del contrato
	contract := bind.NewBoundContract(contractAddr, parsedABI, bs.Client, bs.Client, bs.Client)
	callOpts := &bind.CallOpts{Context: context.Background()}
	
	// Intentar llamar a cada función de solo lectura
	functions := []string{"loteId", "fabricante", "propietarioActual", "temperaturaMinima", "temperaturaMaxima", "comprometido"}
	
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