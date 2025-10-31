package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	AlchemyAPIKey string
	AlchemyWSURL  string
	Port          string
}

func LoadConfig() *Config {
	// Cargar .env si existe
	if err := godotenv.Load(); err != nil {
		log.Println("No se encontró archivo .env, usando variables de entorno del sistema")
	}

	config := &Config{
		AlchemyAPIKey: getEnv("ALCHEMY_API_KEY", ""),
		AlchemyWSURL:  getEnv("ALCHEMY_WS_URL", "wss://eth-sepolia.g.alchemy.com/v2"),
		Port:          getEnv("PORT", "8081"),
	}

	if config.AlchemyAPIKey == "" {
		log.Fatal("ALCHEMY_API_KEY es requerida")
	}

	return config
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}