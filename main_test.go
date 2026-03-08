package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func newTestApp(t *testing.T) *App {
	t.Helper()

	dataDir := t.TempDir()
	t.Setenv("APP_MESSAGE", "test message")
	t.Setenv("DATA_DIR", dataDir)

	return NewApp()
}

func TestHealth(t *testing.T) {
	app := newTestApp(t)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	app.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode json: %v", err)
	}

	if body["status"] != "ok" {
		t.Fatalf("expected status ok, got %q", body["status"])
	}

	if body["message"] != "test message" {
		t.Fatalf("expected message test message, got %q", body["message"])
	}
}

func TestIndexShowsMessage(t *testing.T) {
	app := newTestApp(t)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	app.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "test message") {
		t.Fatalf("expected body to contain app message, got %q", body)
	}

	if !strings.Contains(body, "Visits: 0") {
		t.Fatalf("expected body to contain visit count 0, got %q", body)
	}
}

func TestVisitPersistsCount(t *testing.T) {
	app := newTestApp(t)

	req1 := httptest.NewRequest(http.MethodPost, "/visit", nil)
	rec1 := httptest.NewRecorder()
	app.ServeHTTP(rec1, req1)

	if rec1.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec1.Code)
	}

	req2 := httptest.NewRequest(http.MethodPost, "/visit", nil)
	rec2 := httptest.NewRecorder()
	app.ServeHTTP(rec2, req2)

	if rec2.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec2.Code)
	}

	data, err := os.ReadFile(app.visitsFile())
	if err != nil {
		t.Fatalf("failed to read visits file: %v", err)
	}

	if strings.TrimSpace(string(data)) != "2" {
		t.Fatalf("expected persisted count 2, got %q", string(data))
	}
}

func TestVisitMethodNotAllowed(t *testing.T) {
	app := newTestApp(t)

	req := httptest.NewRequest(http.MethodGet, "/visit", nil)
	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}

func TestServeHTTPLogsRequest(t *testing.T) {
	app := newTestApp(t)

	var buf bytes.Buffer
	originalWriter := log.Writer()
	originalFlags := log.Flags()
	log.SetOutput(&buf)
	log.SetFlags(0)
	defer log.SetOutput(originalWriter)
	defer log.SetFlags(originalFlags)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	req.RemoteAddr = "127.0.0.1:12345"
	rec := httptest.NewRecorder()

	app.ServeHTTP(rec, req)

	output := buf.String()
	if !strings.Contains(output, "http request method=GET path=/health status=200") {
		t.Fatalf("expected access log line, got %q", output)
	}

	if !strings.Contains(output, "remote_addr=127.0.0.1:12345") {
		t.Fatalf("expected remote address in log line, got %q", output)
	}
}

func TestListenAddr(t *testing.T) {
	tests := []struct {
		name string
		host string
		port string
		want string
	}{
		{
			name: "all interfaces",
			host: "0.0.0.0",
			port: "8080",
			want: "0.0.0.0:8080",
		},
		{
			name: "empty host uses wildcard",
			host: "",
			port: "8080",
			want: ":8080",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := listenAddr(tt.host, tt.port)
			if got != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, got)
			}
		})
	}
}
