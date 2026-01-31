package main

import (
	"net/http"
	"strings"

	"github.com/bkwaku/nostos/kafka"
	"github.com/bkwaku/nostos/server"
)

// sets up our server and decouples the config from our server instantiation
func injectServer(configuration ConfigVar) *http.Server {
	brokers := strings.Split(configuration.KafkaBrokers, ",")
	producer := kafka.NewProducer(brokers, configuration.KafkaTopic)
	srv := server.NewServer(producer)
	mux := http.NewServeMux()
	srv.RegisterRoutes(mux)
	httpServer := &http.Server{
		Addr:    configuration.ServerAddr,
		Handler: mux,
	}
	return httpServer
}
