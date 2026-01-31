package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	config := LoadConfig()
	httpServer := initilizeServer(config)

	// set up showdown gracefully
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		log.Println("[ingress] shutting down...")

		ctxTimeout, cancelTimeout := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancelTimeout()

		if err := httpServer.Shutdown(ctxTimeout); err != nil {
			log.Printf("[ingress] error shutting down server: %v\n", err)
		}
	}()

	// Start the server
	fmt.Printf("[ingress] listening on %s\n", config.ServerAddr)
	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("HTTP server error: %v", err)
	}
}
