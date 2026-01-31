package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleIngest(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		wantStatusCode int
		wantBody       string
	}{
		{
			name:           "POST returns success",
			method:         http.MethodPost,
			wantStatusCode: http.StatusOK,
			wantBody:       "Ingest endpoint hit\n",
		},
		{
			name:           "GET returns method not allowed",
			method:         http.MethodGet,
			wantStatusCode: http.StatusMethodNotAllowed,
			wantBody:       "Method not allowed\n",
		},
		{
			name:           "PUT returns method not allowed",
			method:         http.MethodPut,
			wantStatusCode: http.StatusMethodNotAllowed,
			wantBody:       "Method not allowed\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := NewServer()
			rr := httptest.NewRecorder()
			req := httptest.NewRequest(tt.method, "/ingest", nil)

			srv.handleIngest(rr, req)

			if rr.Code != tt.wantStatusCode {
				t.Errorf("status code: got %d, want %d", rr.Code, tt.wantStatusCode)
			}

			if rr.Body.String() != tt.wantBody {
				t.Errorf("body: got %q, want %q", rr.Body.String(), tt.wantBody)
			}
		})
	}
}
