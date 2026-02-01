package server

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// Mock Producer for testing
type mockProducer struct {
	sendFunc  func(ctx context.Context, key string, value []byte) error
	sendCalls []sendCall
}

type sendCall struct {
	ctx   context.Context
	key   string
	value []byte
}

func (m *mockProducer) Send(ctx context.Context, key string, value []byte) error {
	m.sendCalls = append(m.sendCalls, sendCall{ctx: ctx, key: key, value: value})
	if m.sendFunc != nil {
		return m.sendFunc(ctx, key, value)
	}
	return nil
}

func TestHandleIngest_Success(t *testing.T) {
	mock := &mockProducer{}
	server := NewServer(mock)

	payload := `{"test": "data"}`
	req := httptest.NewRequest(http.MethodPost, "/ingest", strings.NewReader(payload))
	w := httptest.NewRecorder()

	server.handleIngest(w, req)

	if w.Code != http.StatusAccepted {
		t.Errorf("expected status %d, got %d", http.StatusAccepted, w.Code)
	}

	var response map[string]string
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response["job_id"] == "" {
		t.Error("expected job_id in response")
	}

	if len(mock.sendCalls) != 1 {
		t.Errorf("expected 1 producer call, got %d", len(mock.sendCalls))
	}

	if len(mock.sendCalls) > 0 {
		var message JobMessage
		if err := json.Unmarshal(mock.sendCalls[0].value, &message); err != nil {
			t.Fatalf("failed to unmarshal sent message: %v", err)
		}

		if message.JobID == "" {
			t.Error("expected job_id in message")
		}
		var payloadData map[string]any
		if err := json.Unmarshal(message.Payload, &payloadData); err != nil {
			t.Fatalf("failed to unmarshal payload: %v", err)
		}
		if payloadData["test"] != "data" {
			t.Errorf("expected payload to contain test:data, got %v", payloadData)
		}

		if message.ReceivedAt.IsZero() {
			t.Error("expected received_at to be set")
		}
	}
}

func TestHandleIngest_MethodNotAllowed(t *testing.T) {
	mock := &mockProducer{}
	server := NewServer(mock)

	methods := []string{http.MethodGet, http.MethodPut, http.MethodDelete, http.MethodPatch}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/ingest", nil)
			w := httptest.NewRecorder()

			server.handleIngest(w, req)

			if w.Code != http.StatusMethodNotAllowed {
				t.Errorf("expected status %d, got %d", http.StatusMethodNotAllowed, w.Code)
			}

			if len(mock.sendCalls) != 0 {
				t.Error("producer should not be called for non-POST requests")
			}
		})
	}
}

func TestHandleIngest_LargePayload(t *testing.T) {
	mock := &mockProducer{}
	server := NewServer(mock)

	largePayload := make([]byte, server.maxPayload+1)
	for i := range largePayload {
		largePayload[i] = 'a'
	}

	req := httptest.NewRequest(http.MethodPost, "/ingest", bytes.NewReader(largePayload))
	w := httptest.NewRecorder()

	server.handleIngest(w, req)

	if w.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("expected status %d, got %d", http.StatusRequestEntityTooLarge, w.Code)
	}

	if len(mock.sendCalls) != 0 {
		t.Error("producer should not be called for oversized payload")
	}
}

func TestBuildMessage_InvalidJSON(t *testing.T) {
	server := NewServer(&mockProducer{})
	jobID := "test-job-123"
	body := []byte(`{invalid json}`)

	_, err := server.buildMessage(jobID, body)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestBuildMessage(t *testing.T) {
	server := NewServer(&mockProducer{})
	jobID := "test-job-123"
	body := []byte(`{"test": "data"}`)

	message, err := server.buildMessage(jobID, body)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if message.JobID != jobID {
		t.Errorf("expected job_id %q, got %q", jobID, message.JobID)
	}

	var payload map[string]any
	if err := json.Unmarshal(message.Payload, &payload); err != nil {
		t.Fatalf("failed to unmarshal payload: %v", err)
	}

	if payload["test"] != "data" {
		t.Errorf("expected payload to contain test data")
	}

	if message.ReceivedAt.IsZero() {
		t.Error("expected received_at to be set")
	}
}

func TestIngestIntoQueue_Success(t *testing.T) {
	mock := &mockProducer{}
	server := NewServer(mock)
	logger := slog.Default()

	message := JobMessage{
		JobID:      "test-123",
		Payload:    json.RawMessage(`{"test": "data"}`),
		ReceivedAt: time.Now().UTC(),
	}

	err := server.ingestIntoQueue(context.Background(), "test-123", message, logger)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(mock.sendCalls) != 1 {
		t.Errorf("expected 1 send call, got %d", len(mock.sendCalls))
	}

	if mock.sendCalls[0].key != "test-123" {
		t.Errorf("expected key %q, got %q", "test-123", mock.sendCalls[0].key)
	}
}

func TestIngestIntoQueue_ProducerError(t *testing.T) {
	expectedErr := errors.New("producer error")
	mock := &mockProducer{
		sendFunc: func(ctx context.Context, key string, value []byte) error {
			return expectedErr
		},
	}
	server := NewServer(mock)
	logger := slog.Default()

	message := JobMessage{
		JobID:      "test-123",
		Payload:    json.RawMessage(`{"test": "data"}`),
		ReceivedAt: time.Now().UTC(),
	}

	err := server.ingestIntoQueue(context.Background(), "test-123", message, logger)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "send to producer") {
		t.Errorf("expected wrapped error message, got: %v", err)
	}
}

func TestWriteResponseToClient(t *testing.T) {
	server := NewServer(&mockProducer{})
	logger := slog.Default()
	resp := IngestResponse{JobID: "test-job-456"}

	w := httptest.NewRecorder()
	server.writeResponseToClient(w, resp, logger)

	if w.Code != http.StatusAccepted {
		t.Errorf("expected status %d, got %d", http.StatusAccepted, w.Code)
	}

	if w.Header().Get("Content-Type") != "application/json" {
		t.Errorf("expected Content-Type application/json, got %q", w.Header().Get("Content-Type"))
	}

	var response IngestResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.JobID != resp.JobID {
		t.Errorf("expected job_id %q, got %q", resp.JobID, response.JobID)
	}
}

func TestRegisterRoutes(t *testing.T) {
	server := NewServer(&mockProducer{})

	mux := http.NewServeMux()
	server.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodPost, "/ingest", strings.NewReader("test"))
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code == http.StatusNotFound {
		t.Error("expected /ingest route to be registered")
	}
}

type errorReader struct{}

func (e *errorReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("read error")
}
