package main

import (
	"fmt"
	"net/http"
)

type Server struct {
}

// Acts as a constructor for Server will return just a pointer to Server
func NewServer() *Server {
	return &Server{}
}

func (s *Server) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/ingest", s.handleIngest)
}
func (s *Server) handleIngest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	fmt.Fprintln(w, "Ingest endpoint hit")
}
