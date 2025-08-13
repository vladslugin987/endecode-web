package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Port         string
	Environment  string
	LogLevel     string
	AllowedOrigins string
	APIToken     string
	
	// Database
	DBHost     string
	DBPort     string
	DBName     string
	DBUser     string
	DBPassword string
	
	// Redis
	RedisHost string
	RedisPort string
	
	// File Processing
	PhotosPath    string
	ProcessedPath string
	TempPath      string
	UploadsPath   string
	MaxFileSize   string
	WorkerCount   int
	
	// Watermarks
	WatermarkEnabled       bool
	SwapEnabled           bool
	VisibleWatermarkEnabled bool
	
	// Email
	SMTPHost     string
	SMTPPort     string
	SMTPUser     string
	SMTPPassword string
	AdminEmail   string
	
	// Telegram
	TelegramToken  string
	TelegramChatID string
}

func Load() *Config {
	// Try to load .env file if it exists
	godotenv.Load()
	
	workerCount, _ := strconv.Atoi(getEnv("WORKER_COUNT", "4"))
	
	return &Config{
		Port:        getEnv("PORT", "8080"),
		Environment: getEnv("ENVIRONMENT", "development"),
		LogLevel:    getEnv("LOG_LEVEL", "info"),
		AllowedOrigins: getEnv("ALLOWED_ORIGINS", "http://localhost:8080"),
		APIToken:    getEnv("API_TOKEN", ""),
		
		// Database
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBName:     getEnv("DB_NAME", "photoprocessing"),
		DBUser:     getEnv("DB_USER", "photoprocessing"),
		DBPassword: getEnv("DB_PASSWORD", "password"),
		
		// Redis
		RedisHost: getEnv("REDIS_HOST", "localhost"),
		RedisPort: getEnv("REDIS_PORT", "6379"),
		
		// File Processing
		PhotosPath:    getEnv("PHOTOS_PATH", "/app/data/photos"),
		ProcessedPath: getEnv("PROCESSED_PATH", "/app/data/processed"),
		TempPath:      getEnv("TEMP_PATH", "/app/data/temp"),
		UploadsPath:   getEnv("UPLOADS_PATH", "/app/uploads"),
		MaxFileSize:   getEnv("MAX_FILE_SIZE", "100MB"),
		WorkerCount:   workerCount,
		
		// Watermarks
		WatermarkEnabled:        getBoolEnv("WATERMARK_ENABLED", true),
		SwapEnabled:            getBoolEnv("SWAP_ENABLED", true),
		VisibleWatermarkEnabled: getBoolEnv("VISIBLE_WATERMARK_ENABLED", true),
		
		// Email
		SMTPHost:     getEnv("SMTP_HOST", ""),
		SMTPPort:     getEnv("SMTP_PORT", "587"),
		SMTPUser:     getEnv("SMTP_USER", ""),
		SMTPPassword: getEnv("SMTP_PASSWORD", ""),
		AdminEmail:   getEnv("ADMIN_EMAIL", ""),
		
		// Telegram
		TelegramToken:  getEnv("TELEGRAM_TOKEN", ""),
		TelegramChatID: getEnv("TELEGRAM_CHAT_ID", ""),
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getBoolEnv(key string, defaultValue bool) bool {
	if value, exists := os.LookupEnv(key); exists {
		if parsed, err := strconv.ParseBool(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func (c *Config) RedisAddr() string {
	return c.RedisHost + ":" + c.RedisPort
}