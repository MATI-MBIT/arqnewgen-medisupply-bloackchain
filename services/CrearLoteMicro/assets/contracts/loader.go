package contracts

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"
)

// Archivos embebidos en el binario
//go:embed LoteTracing.abi.json
var loteTracingABI string

//go:embed LoteTracing.bytecode
var loteTracingBytecode string

//go:embed contract_info.json
var contractInfoJSON string

// ContractInfo contiene información del contrato
type ContractInfo struct {
	ContractName string   `json:"contractName"`
	Version      string   `json:"version"`
	Compiler     string   `json:"compiler"`
	LastUpdated  string   `json:"lastUpdated"`
	Source       string   `json:"source"`
	Hash         string   `json:"hash"`
	Description  string   `json:"description"`
	Features     []string `json:"features"`
	Events       []string `json:"events"`
	Functions    []string `json:"functions"`
}

// GetLoteTracingABI retorna el ABI del contrato LoteTracing
func GetLoteTracingABI() string {
	return strings.TrimSpace(loteTracingABI)
}

// GetLoteTracingBytecode retorna el bytecode del contrato LoteTracing
func GetLoteTracingBytecode() string {
	return strings.TrimSpace(loteTracingBytecode)
}

// GetContractInfo retorna la información del contrato
func GetContractInfo() (*ContractInfo, error) {
	var info ContractInfo
	err := json.Unmarshal([]byte(contractInfoJSON), &info)
	if err != nil {
		return nil, fmt.Errorf("error parsing contract info: %v", err)
	}
	return &info, nil
}

// ValidateContract verifica que los assets del contrato estén disponibles
func ValidateContract() error {
	if strings.TrimSpace(loteTracingABI) == "" {
		return fmt.Errorf("ABI del contrato LoteTracing no está disponible")
	}
	
	if strings.TrimSpace(loteTracingBytecode) == "" {
		return fmt.Errorf("Bytecode del contrato LoteTracing no está disponible")
	}
	
	if !strings.HasPrefix(strings.TrimSpace(loteTracingBytecode), "0x") {
		return fmt.Errorf("Bytecode del contrato tiene formato inválido")
	}
	
	_, err := GetContractInfo()
	if err != nil {
		return fmt.Errorf("información del contrato no válida: %v", err)
	}
	
	return nil
}