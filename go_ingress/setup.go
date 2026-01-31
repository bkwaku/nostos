package main

import "net/http"

// sets up our server and decouples the config from our server creation
func initilizeServer(configuration ConfigVar) *http.Server {
	srv := NewServer()
	mux := http.NewServeMux()
	srv.RegisterRoutes(mux)
	httpServer := &http.Server{
		Addr:    configuration.ServerAddr,
		Handler: mux,
	}
	return httpServer
}
