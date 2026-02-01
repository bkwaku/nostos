package server

import (
	"context"
	"encoding/json"
	"time"
)

type Producer interface {
	Send(ctx context.Context, key string, value []byte) error
}

type IngestResponse struct {
	JobID string `json:"job_id"`
}

type JobMessage struct {
	JobID      string          `json:"job_id"`
	Payload    json.RawMessage `json:"payload"`
	ReceivedAt time.Time       `json:"received_at"`
}

type requestIDKeyType struct{}

var requestIDKey requestIDKeyType

func getRequestID(ctx context.Context) string {
	if id, ok := ctx.Value(requestIDKey).(string); ok {
		return id
	}
	return "unknown"
}
