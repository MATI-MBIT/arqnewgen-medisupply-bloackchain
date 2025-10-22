package utils

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

// ABI del contrato LoteTracing para decodificación - Compilado con Hardhat
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

// DecodedTransaction representa una transacción decodificada
type DecodedTransaction struct {
	FunctionName   string                 `json:"functionName"`
	FunctionSig    string                 `json:"functionSig"`
	Parameters     map[string]interface{} `json:"parameters"`
	RawInputData   string                 `json:"rawInputData"`
}

// DecodeInputData decodifica el input data de una transacción
func DecodeInputData(inputData string) (*DecodedTransaction, error) {
	// Remover el prefijo 0x si existe
	if strings.HasPrefix(inputData, "0x") {
		inputData = inputData[2:]
	}

	// Verificar que tenga al menos 8 caracteres (4 bytes para el selector)
	if len(inputData) < 8 {
		return nil, fmt.Errorf("input data demasiado corto")
	}

	// Extraer el function selector (primeros 4 bytes)
	functionSelector := inputData[:8]
	parametersData := inputData[8:]

	// Parsear el ABI
	parsedABI, err := abi.JSON(strings.NewReader(contractABI))
	if err != nil {
		return nil, fmt.Errorf("error parseando ABI: %v", err)
	}

	// Buscar la función por su selector
	var foundMethod *abi.Method
	var methodName string

	for name, method := range parsedABI.Methods {
		methodID := method.ID
		if fmt.Sprintf("%x", methodID) == functionSelector {
			foundMethod = &method
			methodName = name
			break
		}
	}

	if foundMethod == nil {
		return &DecodedTransaction{
			FunctionName: "Unknown",
			FunctionSig:  "0x" + functionSelector,
			Parameters:   map[string]interface{}{"raw": "0x" + parametersData},
			RawInputData: "0x" + inputData,
		}, nil
	}

	// Decodificar los parámetros
	parameters := make(map[string]interface{})
	
	if len(parametersData) > 0 {
		// Convertir hex a bytes
		paramBytes, err := hexutil.Decode("0x" + parametersData)
		if err != nil {
			return nil, fmt.Errorf("error decodificando parámetros hex: %v", err)
		}

		// Desempaquetar los parámetros usando el ABI
		values, err := foundMethod.Inputs.Unpack(paramBytes)
		if err != nil {
			return nil, fmt.Errorf("error desempaquetando parámetros: %v", err)
		}

		// Mapear los valores a los nombres de parámetros
		for i, input := range foundMethod.Inputs {
			if i < len(values) {
				parameters[input.Name] = formatValue(values[i])
			}
		}
	}

	return &DecodedTransaction{
		FunctionName: methodName,
		FunctionSig:  foundMethod.Sig,
		Parameters:   parameters,
		RawInputData: "0x" + inputData,
	}, nil
}

// formatValue formatea un valor para que sea más legible
func formatValue(value interface{}) interface{} {
	switch v := value.(type) {
	case common.Address:
		return v.Hex()
	case *big.Int:
		return v.String()
	case []byte:
		return hexutil.Encode(v)
	case int8:
		return v
	case uint8:
		return v
	case bool:
		return v
	case string:
		return v
	default:
		return fmt.Sprintf("%v", v)
	}
}

// DecodeSpecificInputData decodifica el input data específico que proporcionaste
func DecodeSpecificInputData() (*DecodedTransaction, error) {
	inputData := "0xf7b5b4e90000000000000000000000000000000000000000000000000000000000000006"
	return DecodeInputData(inputData)
}

// GetFunctionSignatures retorna todas las signatures de funciones del contrato
func GetFunctionSignatures() map[string]string {
	parsedABI, err := abi.JSON(strings.NewReader(contractABI))
	if err != nil {
		return nil
	}

	signatures := make(map[string]string)
	for name, method := range parsedABI.Methods {
		signatures[fmt.Sprintf("%x", method.ID)] = fmt.Sprintf("%s(%s)", name, method.Sig)
	}

	return signatures
}