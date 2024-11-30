package config

import (
    "os"
    "strconv"
    "time"
)

type Config struct {
    RedisAddr        string
    RedisPassword    string
    RedisDB         int
    ServerPort      string
    ReservationTTL  time.Duration
}

func LoadConfig() *Config {
    return &Config{
        RedisAddr:       getEnv("REDIS_ADDR", "localhost:6379"),
        RedisPassword:   getEnv("REDIS_PASSWORD", ""),
        RedisDB:        getEnvAsInt("REDIS_DB", 0),
        ServerPort:     getEnv("SERVER_PORT", "8080"),
        ReservationTTL: getEnvAsDuration("RESERVATION_TTL", 30*time.Minute),
    }
}

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
    if value := os.Getenv(key); value != "" {
        if intVal, err := strconv.Atoi(value); err == nil {
            return intVal
        }
    }
    return defaultValue
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
    if value := os.Getenv(key); value != "" {
        if duration, err := time.ParseDuration(value); err == nil {
            return duration
        }
    }
    return defaultValue
}