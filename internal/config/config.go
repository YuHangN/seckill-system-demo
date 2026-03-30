// Package config internal/config/config.go
package config

import "os"

type Config struct {
	KafkaBrokers string
	RedisAddr    string
	DBDSN        string
	Port         string
}

// Load reads environment variables. Panics if required vars are missing.
func Load() Config {
	c := Config{
		KafkaBrokers: getEnv("KAFKA_BROKERS", "localhost:9092"),
		RedisAddr:    getEnv("REDIS_ADDR", "localhost:6379"),
		DBDSN:        getEnv("DB_DSN", "host=localhost user=postgres password=postgres dbname=seckill sslmode=disable"),
		Port:         getEnv("PORT", "8080"),
	}
	return c
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
