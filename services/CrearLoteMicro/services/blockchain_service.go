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

// ABI del contrato LoteTracing - Actualizado con nuevos campos y funciones
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
			{"indexed": true, "internalType": "address", "name": "nuevoPropietario", "type": "address"},
			{"indexed": false, "internalType": "bool", "name": "comprometido", "type": "bool"}
		],
		"name": "CustodiaTransferida",
		"type": "event"
	},
	{
		"anonymous": false,
		"inputs": [
			{"indexed": true, "internalType": "address", "name": "propietario", "type": "address"},
			{"indexed": false, "internalType": "int8", "name": "tempMin", "type": "int8"},
			{"indexed": false, "internalType": "int8", "name": "tempMax", "type": "int8"},
			{"indexed": false, "internalType": "bool", "name": "comprometido", "type": "bool"},
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
			{"internalType": "int8", "name": "_tempMin", "type": "int8"},
			{"internalType": "int8", "name": "_tempMax", "type": "int8"}
		],
		"name": "registrarTemperatura",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [],
		"name": "tempRegMaxima",
		"outputs": [{"internalType": "int8", "name": "", "type": "int8"}],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [],
		"name": "tempRegMinima",
		"outputs": [{"internalType": "int8", "name": "", "type": "int8"}],
		"stateMutability": "view",
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

// Bytecode del contrato (necesario para el deploy) - Actualizado con la versión final
const contractBytecode = "0x60e060405234801561000f575f5ffd5b506040516109e33803806109e383398101604081905261002e916100f2565b5f6100398482610246565b50336080819052600180545f85810b60a05284900b60c0526001600160b81b03191662010000830261ffff60ff60b01b01191617905560405161007d908590610300565b604080519182900382205f86810b845285900b6020840152917f614b7bdb598a394ea5f748900a1d99d9db2adc2a0662e25111fbf0b6d06dbf74910160405180910390a3505050610316565b634e487b7160e01b5f52604160045260245ffd5b80515f81900b81146100ed575f5ffd5b919050565b5f5f5f60608486031215610104575f5ffd5b83516001600160401b03811115610119575f5ffd5b8401601f81018613610129575f5ffd5b80516001600160401b03811115610142576101426100c9565b604051601f8201601f19908116603f011681016001600160401b0381118282101715610170576101706100c9565b604052818152828201602001881015610187575f5ffd5b8160208401602083015e5f602083830101528095505050506101ab602085016100dd565b91506101b9604085016100dd565b90509250925092565b600181811c908216806101d657607f821691505b6020821081036101f457634e487b7160e01b5f52602260045260245ffd5b50919050565b601f82111561024157805f5260205f20601f840160051c8101602085101561021f5750805b601f840160051c820191505b8181101561023e575f815560010161022b565b50505b505050565b81516001600160401b0381111561025f5761025f6100c9565b6102738161026d84546101c2565b846101fa565b6020601f8211600181146102a5575f831561028e5750848201515b5f19600385901b1c1916600184901b17845561023e565b5f84815260208120601f198516915b828110156102d457878501518255602094850194600190920191016102b4565b50848210156102f157868401515f19600387901b60f8161c191681555b50505050600190811b01905550565b5f82518060208501845e5f920191825250919050565b60805160a05160c05161068861035b5f395f818160cc015281816102b601526102e401525f81816101a30152818161025a015261028801525f61011601526106885ff3fe608060405234801561000f575f5ffd5b506004361061009b575f3560e01c806386b7d1e01161006357806386b7d1e014610150578063902e6d661461017457806395defb5614610185578063af1e62531461019e578063d48cf490146101c5575f5ffd5b80630bf3a8631461009f5780631ccbe36b146100b45780632ba6b752146100c75780633f3a74a41461010557806346ed76f114610111575b5f5ffd5b6100b26100ad366004610587565b6101da565b005b6100b26100c23660046105b8565b6103ae565b6100ee7f000000000000000000000000000000000000000000000000000000000000000081565b6040515f9190910b81526020015b60405180910390f35b6001546100ee905f0b81565b6101387f000000000000000000000000000000000000000000000000000000000000000081565b6040516001600160a01b0390911681526020016100fc565b60015461016490600160b01b900460ff1681565b60405190151581526020016100fc565b6001546100ee9061010090045f0b81565b600154610138906201000090046001600160a01b031681565b6100ee7f000000000000000000000000000000000000000000000000000000000000000081565b6101cd6104e7565b6040516100fc91906105e5565b600154600160b01b900460ff16156102395760405162461bcd60e51b815260206004820152601c60248201527f456c206c6f7465207961206573746120636f6d70726f6d657469646f0000000060448201526064015b60405180910390fd5b6001805460ff8381166101000261ffff19909216908516171790555f82810b7f000000000000000000000000000000000000000000000000000000000000000090910b12806102ab5750805f0b7f00000000000000000000000000000000000000000000000000000000000000005f0b135b806102d95750815f0b7f00000000000000000000000000000000000000000000000000000000000000005f0b125b806103075750805f0b7f00000000000000000000000000000000000000000000000000000000000000005f0b135b156103aa576001805460ff60b01b1916600160b01b90811791829055604080515f86810b825285900b60208201529190920460ff16151591810191909152608060608201819052601a908201527f54656d70657261747572612066756572612064652072616e676f00000000000060a082015233907f26174a1d6632f37659819648fe37603b6961a9c4af42e0dc262b371b8320d76b9060c00160405180910390a25b5050565b6001546201000090046001600160a01b031633146104275760405162461bcd60e51b815260206004820152603060248201527f416363696f6e20736f6c6f207065726d6974696461207061726120656c20707260448201526f1bdc1a595d185c9a5bc81858dd1d585b60821b6064820152608401610230565b6001600160a01b0381166104725760405162461bcd60e51b8152602060048201526012602482015271446972656363696f6e20696e76616c69646160701b6044820152606401610230565b600180546001600160a01b038381166201000081810262010000600160b01b03198516179485905560405160ff600160b01b9096049590951615158552909204169182907f0fef6771eca134aaa0a42e1d6a5e8fcbd185e2148341f7158fd3460de9fc2c6b9060200160405180910390a35050565b5f80546104f39061061a565b80601f016020809104026020016040519081016040528092919081815260200182805461051f9061061a565b801561056a5780601f106105415761010080835404028352916020019161056a565b820191905f5260205f20905b81548152906001019060200180831161054d57829003601f168201915b505050505081565b80355f81900b8114610582575f5ffd5b919050565b5f5f60408385031215610598575f5ffd5b6105a183610572565b91506105af60208401610572565b90509250929050565b5f602082840312156105c8575f5ffd5b81356001600160a01b03811681146105de575f5ffd5b9392505050565b602081525f82518060208401528060208501604085015e5f604082850101526040601f19601f83011684010191505092915050565b600181811c9082168061062e57607f821691505b60208210810361064c57634e487b7160e01b5f52602260045260245ffd5b5091905056fea26469706673582212203e3b19841698ac48ca1b569a0ed3b28fd4ab11a5c3010c60c48e32d94bdacb5d64736f6c634300081c0033"

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
	parsedABI, err := abi.JSON(strings.NewReader(contractABI))
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
	parsedABI, err := abi.JSON(strings.NewReader(contractABI))
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