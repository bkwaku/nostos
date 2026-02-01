package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type Server struct {
	logger     *slog.Logger
	producer   Producer
	maxPayload int64
	timeout    time.Duration
}

// Acts as a constructor for Server will return just a pointer to Server
func NewServer(producer Producer) *Server {
	return &Server{
		maxPayload: 1 << 20, // enforce Kafka recommendation for max message size
		logger:     slog.Default(),
		producer:   producer,
		timeout:    5 * time.Second,
	}
}

// all request must hit this endpoint
func (s *Server) RegisterRoutes(mux *http.ServeMux) {
	handler := http.HandlerFunc(s.handleIngest)
	// a small middleware to log requests
	mux.Handle("/ingest", s.loggingMiddleware(handler))
}

func (s *Server) handleIngest(w http.ResponseWriter, r *http.Request) {
	requestID := getRequestID(r.Context())
	logger := s.logger.With(slog.String("request_id", requestID))

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// ensure body is closed after reading
	defer r.Body.Close()
	r.Body = http.MaxBytesReader(w, r.Body, s.maxPayload)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			logger.Warn("request body exceeds max payload size",
				slog.Int64("max_payload", s.maxPayload))
			http.Error(w, http.StatusText(http.StatusRequestEntityTooLarge), http.StatusRequestEntityTooLarge)
			return
		}
		// somee unexpected error
		logger.Error("failed to read request body",
			slog.String("error", fmt.Sprintf("%v", err)))
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	jobID := uuid.NewString()
	logger = logger.With(slog.String("job_id", jobID))

	message, err := s.buildMessage(jobID, body)
	if err != nil {
		logger.Warn("invalid JSON in request body",
			slog.String("error", err.Error()))
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	if err := s.ingestIntoQueue(r.Context(), jobID, message, logger); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	resp := IngestResponse{JobID: jobID}
	s.writeResponseToClient(w, resp, logger)
}

func (s *Server) buildMessage(jobID string, body []byte) (JobMessage, error) {
	var payload json.RawMessage
	if err := json.Unmarshal(body, &payload); err != nil {
		return JobMessage{}, fmt.Errorf("invalid JSON payload: %w", err)
	}
	msg := JobMessage{
		JobID:      jobID,
		Payload:    payload,
		ReceivedAt: time.Now().UTC(),
	}
	return msg, nil
}

func (s *Server) ingestIntoQueue(ctx context.Context, jobID string, message JobMessage, logger *slog.Logger) error {
	msgBytes, err := json.Marshal(message)
	if err != nil {
		logger.Error("failed to marshal job message",
			slog.String("error", err.Error()))
		return fmt.Errorf("marshal job message: %w", err)
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if err := s.producer.Send(timeoutCtx, jobID, msgBytes); err != nil {
		logger.Error("failed to send message to producer",
			slog.String("error", err.Error()))
		return fmt.Errorf("send to producer: %w", err)
	}

	logger.Info("message successfully enqueued")
	return nil
}

// the client gets a 202 Accepted with the job ID in the response body
func (s *Server) writeResponseToClient(w http.ResponseWriter, resp IngestResponse, logger *slog.Logger) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		logger.Error("failed to encode response",
			slog.String("error", err.Error()))
		return
	}
}

func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.NewString()
		}

		reqLogger := s.logger.With(
			slog.String("request_id", requestID),
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.String("remote_addr", r.RemoteAddr),
		)

		ctx := context.WithValue(r.Context(), requestIDKey, requestID)
		r = r.WithContext(ctx)

		wrapped := &responseWriter{ResponseWriter: w, status: http.StatusOK}

		wrapped.Header().Set("X-Request-ID", requestID)

		next.ServeHTTP(wrapped, r)

		reqLogger.Info("request completed",
			slog.Int("status", wrapped.status),
			slog.Duration("duration", time.Since(start)),
			slog.Int("bytes", wrapped.bytes),
		)
	})
}
