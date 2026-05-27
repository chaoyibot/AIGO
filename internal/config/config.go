package config

import (
	"os"
	"strconv"
)

type Config struct {
	Server    ServerConfig
	Database  DatabaseConfig
	Redis     RedisConfig
	NATS      NATSConfig
	Points    PointsConfig
	Broadcast BroadcastConfig
	Crypto    CryptoConfig
}

type ServerConfig struct {
	Port    int
	Debug   bool
	BaseURL string
}

type DatabaseConfig struct {
	URL string
}

type RedisConfig struct {
	URL string
}

type NATSConfig struct {
	URL string
}

type PointsConfig struct {
	ExchangeRate int64 // 法币:积分 = 1:100
}

type BroadcastConfig struct {
	BasicCost    int64 // 基础广播积分成本
	StandardCost int64
	PremiumCost  int64
}

type CryptoConfig struct {
	SystemKey string // hex-encoded AES-256 key from SYSTEM_CRYPTO_KEY env
}

func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port:    getEnvInt("PORT", 8080),
			Debug:   getEnv("DEBUG", "false") == "true",
			BaseURL: getEnv("BASE_URL", "http://localhost:8080"),
		},
		Database: DatabaseConfig{
			URL: getEnv("DATABASE_URL", "postgres://aigo:aigo@localhost:5432/aigo?sslmode=disable"),
		},
		Redis: RedisConfig{
			URL: getEnv("REDIS_URL", "redis://localhost:6379/0"),
		},
		NATS: NATSConfig{
			URL: getEnv("NATS_URL", "nats://localhost:4222"),
		},
		Points: PointsConfig{
			ExchangeRate: 100,
		},
		Broadcast: BroadcastConfig{
			BasicCost:    500,
			StandardCost: 2000,
			PremiumCost:  5000,
		},
		Crypto: CryptoConfig{
			SystemKey: os.Getenv("SYSTEM_CRYPTO_KEY"),
		},
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}
