package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

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
	MaxConnections int
	MaxIdleConns   int
	MigrationsPath string
	
	// Redis
	RedisURL      string
	RedisPassword string
	
	// JWT
	JWTSecret string
	JWTExpiry time.Duration
	
	// Midtrans
	MidtransServerKey string
	MidtransClientKey string
	MidtransEnv       string
	
	// AWS S3
	AWSAccessKeyID     string
	AWSSecretAccessKey string
	AWSRegion          string
	AWSBucketName      string
	
	// Email
	SMTPHost string
	SMTPPort int
	SMTPUser string
	SMTPPass string
	
	// WhatsApp
	WAApiURL   string
	WAApiToken string
	
	// Media
	MediaPath string
	
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
		MaxConnections: getEnvAsInt("DB_MAX_CONNECTIONS", 100),
		MaxIdleConns:   getEnvAsInt("DB_MAX_IDLE_CONNECTIONS", 10),
		MigrationsPath: getEnv("MIGRATIONS_PATH", "internal/database/migrations"),
		
		// Redis
		RedisURL:      getEnv("REDIS_URL", "redis://localhost:6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		
		// JWT
		JWTSecret: getEnv("JWT_SECRET", ""),
		JWTExpiry: getEnvAsDuration("JWT_EXPIRY", "24h"),
		
		// Midtrans
		MidtransServerKey: getEnv("MIDTRANS_SERVER_KEY", ""),
		MidtransClientKey: getEnv("MIDTRANS_CLIENT_KEY", ""),
		MidtransEnv:       getEnv("MIDTRANS_ENV", "sandbox"),
		
		// AWS S3
		AWSAccessKeyID:     getEnv("AWS_ACCESS_KEY_ID", ""),
		AWSSecretAccessKey: getEnv("AWS_SECRET_ACCESS_KEY", ""),
		AWSRegion:          getEnv("AWS_REGION", "ap-southeast-1"),
		AWSBucketName:      getEnv("AWS_BUCKET_NAME", ""),
		
		// Email
		SMTPHost: getEnv("SMTP_HOST", "smtp.gmail.com"),
		SMTPPort: getEnvAsInt("SMTP_PORT", 587),
		SMTPUser: getEnv("SMTP_USER", ""),
		SMTPPass: getEnv("SMTP_PASS", ""),
		
		// WhatsApp
		WAApiURL:   getEnv("WA_API_URL", ""),
		WAApiToken: getEnv("WA_API_TOKEN", ""),
		
		// Media
		MediaPath: getEnv("MEDIA_PATH", "uploads"),
		
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
	if c.JWTSecret == "" {
		return fmt.Errorf("JWT_SECRET is required")
	}
	
	if c.MidtransServerKey == "" || c.MidtransClientKey == "" {
		return fmt.Errorf("Midtrans keys are required")
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

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
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

func getEnvAsDuration(key string, defaultValue string) time.Duration {
	valueStr := getEnv(key, defaultValue)
	if duration, err := time.ParseDuration(valueStr); err == nil {
		return duration
	}
	duration, _ := time.ParseDuration(defaultValue)
	return duration
}