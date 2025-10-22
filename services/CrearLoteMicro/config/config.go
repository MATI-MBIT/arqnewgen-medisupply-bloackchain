package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	SepoliaRPC string
	Port       string
	ChainID    int64
}

func LoadConfig() *Config {
	// Cargar variables de entorno desde .env si existe
	if err := godotenv.Load(); err != nil {
		log.Println("No se encontró archivo .env, usando variables de entorno del sistema")
	}

	config := &Config{
		SepoliaRPC: getEnv("SEPOLIA_RPC", "https://sepolia.infura.io/v3/YOUR_PROJECT_ID"),
		Port:       getEnv("PORT", "8080"),
		ChainID:    11155111, // Sepolia Chain ID
	}

	return config
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}