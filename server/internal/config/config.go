package config

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	PGSSLMode        string
	SMTPAPIKey       string
	PGName           string
	SwaggerAddr      string
	PGPassword       string
	PGDB             string
	SQLitePath       string
	PublicAddr       string
	AppName          string
	LogLevel         string
	PGUser           string
	RedisPassword    string
	RedisAddr        string
	JWTRefreshSecret []byte
	JWTAccessSecret  []byte
	RedisDB          int
	PGPort           int
	ShutdownTimeout  int
	IsLocalRun       bool
}

func Load() *Config {
	_ = godotenv.Load()

	accessSecret := getEnvOrGenerateSecret("JWT_ACCESS_SECRET")
	refreshSecret := getEnvOrGenerateSecret("JWT_REFRESH_SECRET")

	return &Config{
		AppName:          "theca",
		LogLevel:         getEnv("LOG_LEVEL", "INFO"),
		PGName:           getEnv("PG_NAME", "postgres"),
		PGUser:           getEnv("PG_USER", "postgres"),
		PGPassword:       getEnv("PG_PASSWORD", "postgres"),
		PGDB:             getEnv("PG_DB", "postgres"),
		PGPort:           getInt("PG_PORT", 5432),
		PGSSLMode:        getEnv("PG_SSL_MODE", "disable"),
		IsLocalRun:       parseBool("IS_LOCAL_RUN"),
		SQLitePath:       getEnv("SQLITE_PATH", "theca_local.db"),
		PublicAddr:       getEnv("PUBLIC_ADDR", ":8080"),
		JWTAccessSecret:  []byte(accessSecret),
		JWTRefreshSecret: []byte(refreshSecret),
		SwaggerAddr:      getEnv("SWAGGER_ADDR", ":8081"),
		SMTPAPIKey:       getEnv("SMTP_API_KEY", ""),
		RedisAddr:        getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword:    getEnv("REDIS_PASSWORD", ""),
		RedisDB:          getInt("REDIS_DB", 0),
		ShutdownTimeout:  getInt("SHUTDOWN_TIMEOUT", 5),
	}
}

func getEnv(key, defaultValue string) string {
	val := os.Getenv(key)
	if val == "" {
		return defaultValue
	}

	return val
}

func parseBool(key string) bool {
	param := os.Getenv(key)
	r := false
	if param != "" {
		var err error
		r, err = strconv.ParseBool(param)
		if err != nil {
			fmt.Printf("WARN: invalid %s value (%s), defualting to false\n", key, param)
		}
	}
	return r
}

func getInt(key string, defaultValue int) int {
	val := os.Getenv(key)
	if val == "" {
		return defaultValue
	}

	intVal, err := strconv.Atoi(val)
	if err != nil {
		return defaultValue
	}
	return intVal
}

func getEnvOrGenerateSecret(key string) string {
	val := os.Getenv(key)
	if val == "" {
		return generateRandomSecret()
	}
	return val
}

func generateRandomSecret() string {
	bytes := make([]byte, 32)
	_, err := rand.Read(bytes)
	if err != nil {
		log.Printf("CRITICAL: Failed to generate cryptographically secure secret: %v", err)
		log.Printf("WARNING: Application security may be compromised. Please set %s environment variable manually.", "JWT secrets")
		panic("Failed to generate secure JWT secret")
	}

	secret := hex.EncodeToString(bytes)
	log.Printf("SECURITY WARNING: Generated random JWT secret. For production, please set environment variables manually.")
	return secret
}
