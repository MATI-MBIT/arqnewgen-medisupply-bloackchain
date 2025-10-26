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

// Bytecode del contrato (necesario para el deploy) - Compatible con Sepolia
const contractBytecode = "0x60e060405234801561001057600080fd5b506040516109ed3803806109ed83398101604081905261002f9161011e565b600061003b848261027b565b5033608081905260018054600085810b60a05284900b60c0526001600160b81b03191662010000830261ffff60ff60b01b011916179055604051610080908590610339565b60408051918290038220600086810b845285900b6020840152917f614b7bdb598a394ea5f748900a1d99d9db2adc2a0662e25111fbf0b6d06dbf74910160405180910390a3505050610355565b634e487b7160e01b600052604160045260246000fd5b60005b838110156100fe5781810151838201526020016100e6565b50506000910152565b8051600081900b811461011957600080fd5b919050565b60008060006060848603121561013357600080fd5b83516001600160401b0381111561014957600080fd5b8401601f8101861361015a57600080fd5b80516001600160401b03811115610173576101736100cd565b604051601f8201601f19908116603f011681016001600160401b03811182821017156101a1576101a16100cd565b6040528181528282016020018810156101b957600080fd5b6101ca8260208301602086016100e3565b94506101db91505060208501610107565b91506101e960408501610107565b90509250925092565b600181811c9082168061020657607f821691505b60208210810361022657634e487b7160e01b600052602260045260246000fd5b50919050565b601f82111561027657806000526020600020601f840160051c810160208510156102535750805b601f840160051c820191505b81811015610273576000815560010161025f565b50505b505050565b81516001600160401b03811115610294576102946100cd565b6102a8816102a284546101f2565b8461022c565b6020601f8211600181146102dc57600083156102c45750848201515b600019600385901b1c1916600184901b178455610273565b600084815260208120601f198516915b8281101561030c57878501518255602094850194600190920191016102ec565b508482101561032a5786840151600019600387901b60f8161c191681555b50505050600190811b01905550565b6000825161034b8184602087016100e3565b9190910192915050565b60805160a05160c05161065c6103916000396000818160d001526102910152600081816101aa01526102620152600061011c015261065c6000f3fe608060405234801561001057600080fd5b506004361061009e5760003560e01c806386b7d1e01161006657806386b7d1e014610156578063902e6d661461017a57806395defb561461018c578063af1e6253146101a5578063d48cf490146101cc57600080fd5b80630bf3a863146100a35780631ccbe36b146100b85780632ba6b752146100cb5780633f3a74a41461010a57806346ed76f114610117575b600080fd5b6100b66100b136600461053b565b6101e1565b005b6100b66100c636600461056e565b61035d565b6100f27f000000000000000000000000000000000000000000000000000000000000000081565b60405160009190910b81526020015b60405180910390f35b6001546100f29060000b81565b61013e7f000000000000000000000000000000000000000000000000000000000000000081565b6040516001600160a01b039091168152602001610101565b60015461016a90600160b01b900460ff1681565b6040519015158152602001610101565b6001546100f290610100900060000b81565b60015461013e906201000090046001600160a01b031681565b6100f27f000000000000000000000000000000000000000000000000000000000000000081565b6101d4610496565b604051610101919061059e565b600154600160b01b900460ff16156102405760405162461bcd60e51b815260206004820152601c60248201527f456c206c6f7465207961206573746120636f6d70726f6d657469646f0000000060448201526064015b60405180910390fd5b6001805460ff8381166101000261ffff1990921690851617179055600082810b7f000000000000000000000000000000000000000000000000000000000000000090910b12806102b557508060000b7f000000000000000000000000000000000000000000000000000000000000000060000b135b15610359576001805460ff60b01b1916600160b01b9081179182905560408051600086810b825285900b60208201529190920460ff16151591810191909152608060608201819052601a908201527f54656d70657261747572612066756572612064652072616e676f00000000000060a082015233907f26174a1d6632f37659819648fe37603b6961a9c4af42e0dc262b371b8320d76b9060c00160405180910390a25b5050565b6001546201000090046001600160a01b031633146103d65760405162461bcd60e51b815260206004820152603060248201527f416363696f6e20736f6c6f207065726d6974696461207061726120656c20707260448201526f1bdc1a595d185c9a5bc81858dd1d585b60821b6064820152608401610237565b6001600160a01b0381166104215760405162461bcd60e51b8152602060048201526012602482015271446972656363696f6e20696e76616c69646160701b6044820152606401610237565b600180546001600160a01b038381166201000081810262010000600160b01b03198516179485905560405160ff600160b01b9096049590951615158552909204169182907f0fef6771eca134aaa0a42e1d6a5e8fcbd185e2148341f7158fd3460de9fc2c6b9060200160405180910390a35050565b600080546104a3906105ec565b80601f01602080910402602001604051908101604052809291908181526020018280546104cf906105ec565b801561051c5780601f106104f15761010080835404028352916020019161051c565b820191906000526020600020905b8154815290600101906020018083116104ff57829003601f168201915b505050505081565b8035600081900b811461053657600080fd5b919050565b6000806040838503121561054e57600080fd5b61055783610524565b915061056560208401610524565b90509250929050565b60006020828403121561058057600080fd5b81356001600160a01b038116811461059757600080fd5b9392505050565b602081526000825180602084015260005b818110156105cc57602081860181015160408684010152016105af565b506000604082850101526040601f19601f83011684010191505092915050565b600181811c9082168061060057607f821691505b60208210810361062057634e487b7160e01b600052602260045260246000fd5b5091905056fea2646970667358221220895abf83133e43dbf014c33b1bd2c368a2487a1da3f2391c0ad8157ca60263fd64736f6c634300081c0033"

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