package config

import (
	"os"
)

type Config struct {
	// Database configuration
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string

	// JWT configuration
	JWTSecret string

	// Server configuration
	Port string

	// MinIO configuration
	MinioEndpoint   string
	MinioAccessKey  string
	MinioSecretKey  string
	MinioBucketName string
	MinioUseSSL     bool

	// Redis configuration
	RedisHost     string
	RedisPort     string
	RedisPassword string
	RedisDB       int
}

var globalConfig *Config

func GetConfig() *Config {
	return globalConfig
}

func LoadConfig() (*Config, error) {
	globalConfig = &Config{
		// Database configuration
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", ""),
		DBName:     getEnv("DB_NAME", "chatroom"),

		// JWT configuration
		JWTSecret: getEnv("JWT_SECRET", "your-secret-key"),

		// Server configuration
		Port: getEnv("PORT", "8080"),

		// MinIO configuration
		MinioEndpoint:   getEnv("MINIO_ENDPOINT", "localhost:9000"),
		MinioAccessKey:  getEnv("MINIO_ACCESS_KEY", "minioadmin"),
		MinioSecretKey:  getEnv("MINIO_SECRET_KEY", "minioadmin"),
		MinioBucketName: getEnv("MINIO_BUCKET_NAME", "avatars"),
		MinioUseSSL:     getEnv("MINIO_USE_SSL", "false") == "true",

		// Redis configuration
		RedisHost:     getEnv("REDIS_HOST", "localhost"),
		RedisPort:     getEnv("REDIS_PORT", "6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		RedisDB:       0,
	}

	return globalConfig, nil
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func SetConfig(cfg *Config) {
	globalConfig = cfg
}
