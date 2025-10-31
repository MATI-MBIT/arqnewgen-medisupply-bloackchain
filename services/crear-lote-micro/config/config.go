package config

import (
	"log"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

type Config struct {
	SepoliaRPC       string
	Port             string
	ChainID          int64
	SepoliaWS        string
	DamageServiceURL string
}

func LoadConfig() *Config {
	// Cargar variables de entorno desde .env si existe
	if err := godotenv.Load(); err != nil {
		log.Println("No se encontró archivo .env, usando variables de entorno del sistema")
	}

	config := &Config{
		SepoliaRPC:       getEnv("SEPOLIA_RPC", "https://sepolia.infura.io/v3/YOUR_PROJECT_ID"),
		Port:             getEnv("PORT", "8080"),
		ChainID:          11155111, // Sepolia Chain ID
		SepoliaWS:        getEnv("SEPOLIA_WS", "wss://eth-sepolia.g.alchemy.com/v2/"+filepath.Base(getEnv("SEPOLIA_RPC", "YOUR_PROJECT_ID"))),
		DamageServiceURL: getEnv("DAMAGE_SERVICE_URL", "http://localhost:8100/contract-broken"),
	}

	return config
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
