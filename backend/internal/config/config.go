package config

import (
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Port                string
	DBHost              string
	DBPort              string
	DBUser              string
	DBPassword          string
	DBName              string
	AIServiceURL        string
	ConfidenceThreshold float64
	HighRiskLevels      []string
	UploadDir           string
	MaxUploadSize       int64
	APIKeySecret        string
}

func Load() *Config {
	godotenv.Load("../.env")

	threshold, _ := strconv.ParseFloat(getEnv("CONFIDENCE_THRESHOLD", "0.8"), 64)
	maxUpload, _ := strconv.ParseInt(getEnv("MAX_UPLOAD_SIZE", "104857600"), 10, 64)

	return &Config{
		Port:                getEnv("BACKEND_PORT", "8080"),
		DBHost:              getEnv("DB_HOST", "localhost"),
		DBPort:              getEnv("DB_PORT", "5432"),
		DBUser:              getEnv("DB_USER", "synapse"),
		DBPassword:          getEnv("DB_PASSWORD", "synapse_dev"),
		DBName:              getEnv("DB_NAME", "synapsechain"),
		AIServiceURL:        getEnv("AI_SERVICE_URL", "http://localhost:8081"),
		ConfidenceThreshold: threshold,
		HighRiskLevels:      strings.Split(getEnv("HIGH_RISK_LEVELS", "high,critical"), ","),
		UploadDir:           getEnv("UPLOAD_DIR", "./uploads"),
		MaxUploadSize:       maxUpload,
		APIKeySecret:        getEnv("API_KEY_SECRET", "change-me-in-production"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
