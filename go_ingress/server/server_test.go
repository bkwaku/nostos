package server

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
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
	server := &Server{
		logger:   log.Default(),
		producer: mock,
	}

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
		var message map[string]any
		if err := json.Unmarshal(mock.sendCalls[0].value, &message); err != nil {
			t.Fatalf("failed to unmarshal sent message: %v", err)
		}

		if message["job_id"] == "" {
			t.Error("expected job_id in message")
		}
		if message["payload"] != payload {
			t.Errorf("expected payload %q, got %q", payload, message["payload"])
		}
		if message["received_at"] == "" {
			t.Error("expected received_at in message")
		}
	}
}

func TestHandleIngest_MethodNotAllowed(t *testing.T) {
	mock := &mockProducer{}
	server := &Server{
		logger:   log.Default(),
		producer: mock,
	}

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
	server := &Server{
		logger:   log.Default(),
		producer: mock,
	}

	largePayload := make([]byte, maxPayloadSize+1)
	for i := range largePayload {
		largePayload[i] = 'a'
	}

	req := httptest.NewRequest(http.MethodPost, "/ingest", bytes.NewReader(largePayload))
	w := httptest.NewRecorder()

	server.handleIngest(w, req)

	if w.Code != http.StatusAccepted {
		t.Errorf("expected status %d, got %d", http.StatusAccepted, w.Code)
	}

	if len(mock.sendCalls) > 0 {
		var message map[string]any
		json.Unmarshal(mock.sendCalls[0].value, &message)
		payload := message["payload"].(string)
		if len(payload) != maxPayloadSize {
			t.Errorf("expected payload size %d, got %d", maxPayloadSize, len(payload))
		}
	}
}

func TestReadPayload_Success(t *testing.T) {
	server := &Server{}
	payload := "test payload"
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(payload))

	body, err := server.readPayload(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if string(body) != payload {
		t.Errorf("expected %q, got %q", payload, string(body))
	}
}

func TestReadPayload_Error(t *testing.T) {
	server := &Server{}
	req := httptest.NewRequest(http.MethodPost, "/", &errorReader{})

	_, err := server.readPayload(req)
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestBuildMessage(t *testing.T) {
	server := &Server{}
	jobID := "test-job-123"
	body := []byte(`{"test": "data"}`)

	message := server.buildMessage(jobID, body)

	if message["job_id"] != jobID {
		t.Errorf("expected job_id %q, got %q", jobID, message["job_id"])
	}

	if message["payload"] != string(body) {
		t.Errorf("expected payload %q, got %q", string(body), message["payload"])
	}

	if message["received_at"] == "" {
		t.Error("expected received_at to be set")
	}
}

func TestProcessIngestIntoQueue_Success(t *testing.T) {
	mock := &mockProducer{}
	server := &Server{producer: mock}

	message := map[string]any{
		"job_id":      "test-123",
		"payload":     "test data",
		"received_at": "2026-01-31T10:00:00Z",
	}

	err := server.processIngestIntoQueue(context.Background(), "test-123", message)
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

func TestWriteResponseToClient(t *testing.T) {
	server := &Server{logger: log.Default()}
	jobID := "test-job-456"

	w := httptest.NewRecorder()
	server.writeResponseToClient(w, jobID)

	if w.Code != http.StatusAccepted {
		t.Errorf("expected status %d, got %d", http.StatusAccepted, w.Code)
	}

	if w.Header().Get("Content-Type") != "application/json" {
		t.Errorf("expected Content-Type application/json, got %q", w.Header().Get("Content-Type"))
	}

	var response map[string]string
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response["job_id"] != jobID {
		t.Errorf("expected job_id %q, got %q", jobID, response["job_id"])
	}
}

func TestRegisterRoutes(t *testing.T) {
	server := &Server{
		logger:   log.Default(),
		producer: &mockProducer{},
	}

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
