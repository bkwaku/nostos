package main

import (
	"os"
)

type ConfigVar struct {
	ServerAddr   string
	KafkaBrokers string
	KafkaTopic   string
}

func LoadConfig() ConfigVar {
	cfg := ConfigVar{}
	cfg.ServerAddr = getEnv("SERVER_ADDR", ":8080")
	cfg.KafkaBrokers = getEnv("KAFKA_BROKERS", "localhost:9092")
	cfg.KafkaTopic = getEnv("KAFKA_TOPIC", "ingress-topic")
	return cfg
}

func getEnv(key, defaultVal string) string {
	v := os.Getenv(key)
	if v == "" {
		return defaultVal
	}
	return v
}
