package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	config := LoadConfig()
	httpServer := injectServer(config)

	// Start the server
	fmt.Printf("[ingress] listening on %s\n", config.ServerAddr)
	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("HTTP server error: %v", err)
	}
}
