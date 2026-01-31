package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// enforce Kafka recommendation for max message size
const maxPayloadSize = 1 << 20 // 1MB

type Producer interface {
	Send(ctx context.Context, key string, value []byte) error
}

type Server struct {
	logger   *log.Logger
	producer Producer
}

// Acts as a constructor for Server will return just a pointer to Server
func NewServer(producer Producer) *Server {
	return &Server{
		logger:   log.Default(),
		producer: producer,
	}
}

// all request must hit this endpoint
func (s *Server) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/ingest", s.handleIngest)
}

func (s *Server) handleIngest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := s.readPayload(r)
	if err != nil {
		s.logger.Printf("failed to read payload: %v", err)
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	jobID := uuid.NewString()
	message := s.buildMessage(jobID, body)

	err = s.processIngestIntoQueue(r.Context(), jobID, message)
	if err != nil {
		s.logger.Printf("failed to process ingest into queue: %v", err)
		http.Error(w, "failed to enqueue request", http.StatusInternalServerError)
		return
	}

	s.writeResponseToClient(w, jobID)
}

func (s *Server) readPayload(r *http.Request) ([]byte, error) {
	body, err := io.ReadAll(io.LimitReader(r.Body, maxPayloadSize))
	if err != nil {
		return nil, err
	}
	return body, nil
}

func (s *Server) buildMessage(jobID string, body []byte) map[string]any {
	return map[string]any{
		"job_id":      jobID,
		"payload":     string(body),
		"received_at": time.Now().UTC().Format(time.RFC3339Nano),
	}
}

func (s *Server) processIngestIntoQueue(ctx context.Context, jobID string, message map[string]any) error {
	msgBytes, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	if err := s.producer.Send(ctx, jobID, msgBytes); err != nil {
		return fmt.Errorf("failed to send to queue: %w", err)
	}

	return nil
}

// the client gets a 202 Accepted with the job ID in the response body
func (s *Server) writeResponseToClient(w http.ResponseWriter, jobID string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	if err := json.NewEncoder(w).Encode(map[string]string{
		"job_id": jobID,
	}); err != nil {
		s.logger.Printf("failed to encode response: %v", err)
	}
}
