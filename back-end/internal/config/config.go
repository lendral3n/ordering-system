package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	// Server
	Port           string
	Env            string
	BaseURL        string
	AllowedOrigins []string
	
	// Database
	DatabaseURL    string
	
	// Midtrans
	MidtransServerKey string
	MidtransClientKey string
	MidtransEnv       string
	
	// Cloudinary
	CloudinaryURL       string
	CloudinaryCloudName string
	CloudinaryAPIKey    string
	CloudinaryAPISecret string
	
	// Tax & Service
	TaxPercentage     float64
	ServicePercentage float64
}

func Load() (*Config, error) {
	// Load .env file if exists
	_ = godotenv.Load()
	
	cfg := &Config{
		// Server
		Port:           getEnv("PORT", "8080"),
		Env:            getEnv("ENV", "development"),
		BaseURL:        getEnv("BASE_URL", "http://localhost:8080"),
		AllowedOrigins: strings.Split(getEnv("ALLOWED_ORIGINS", "*"), ","),
		
		// Database
		DatabaseURL:    getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/restaurant_db?sslmode=disable"),
		
		// Midtrans
		MidtransServerKey: getEnv("MIDTRANS_SERVER_KEY", ""),
		MidtransClientKey: getEnv("MIDTRANS_CLIENT_KEY", ""),
		MidtransEnv:       getEnv("MIDTRANS_ENV", "sandbox"),
		
		// Cloudinary
		CloudinaryURL:       getEnv("CLOUDINARY_URL", ""),
		CloudinaryCloudName: getEnv("CLOUDINARY_CLOUD_NAME", ""),
		CloudinaryAPIKey:    getEnv("CLOUDINARY_API_KEY", ""),
		CloudinaryAPISecret: getEnv("CLOUDINARY_API_SECRET", ""),
		
		// Tax & Service
		TaxPercentage:     getEnvAsFloat("TAX_PERCENTAGE", 10.0),
		ServicePercentage: getEnvAsFloat("SERVICE_PERCENTAGE", 5.0),
	}
	
	// Validate required fields
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	
	return cfg, nil
}

func (c *Config) validate() error {
	if c.MidtransServerKey == "" || c.MidtransClientKey == "" {
		return fmt.Errorf("Midtrans keys are required")
	}
	
	if c.CloudinaryURL == "" {
		return fmt.Errorf("Cloudinary URL is required")
	}
	
	return nil
}

func (c *Config) IsDevelopment() bool {
	return c.Env == "development"
}

func (c *Config) IsProduction() bool {
	return c.Env == "production"
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsFloat(key string, defaultValue float64) float64 {
	valueStr := getEnv(key, "")
	if value, err := strconv.ParseFloat(valueStr, 64); err == nil {
		return value
	}
	return defaultValue
}