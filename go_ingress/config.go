package main

import (
	"os"
)

type ConfigVar struct {
	ServerAddr string
	// Kafka config will go below
}

func LoadConfig() ConfigVar {
	cfg := ConfigVar{}
	cfg.ServerAddr = getEnv("SERVER_ADDR", ":8080")
	return cfg
}

func getEnv(key, defaultVal string) string {
	v := os.Getenv(key)
	if v == "" {
		return defaultVal
	}
	return v
}
