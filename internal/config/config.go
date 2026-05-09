package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port     int
	DBPath   string
	LogLevel string
}

var GlobalConfig Config

func LoadConfig(configPath string) {
	// 1. Set Defaults
	GlobalConfig = Config{
		Port:     8080,
		DBPath:   "/app/data/data.db",
		LogLevel: "info",
	}

	// 2. Override with Environment Variables
	if portStr := os.Getenv("PORT"); portStr != "" {
		if p, err := strconv.Atoi(portStr); err == nil {
			GlobalConfig.Port = p
		}
	}

	if dbPath := os.Getenv("DB_PATH"); dbPath != "" {
		GlobalConfig.DBPath = dbPath
	}

	if logLevel := os.Getenv("LOG_LEVEL"); logLevel != "" {
		GlobalConfig.LogLevel = logLevel
	}
}
