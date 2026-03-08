package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type App struct {
	message string
	dataDir string
	mux     *http.ServeMux
}

func NewApp() *App {
	message := getenv("APP_MESSAGE", "hello from deploy")
	dataDir := getenv("DATA_DIR", "/data")

	app := &App{
		message: message,
		dataDir: dataDir,
		mux:     http.NewServeMux(),
	}

	app.routes()
	return app
}

func (a *App) routes() {
	a.mux.HandleFunc("/", a.handleIndex)
	a.mux.HandleFunc("/health", a.handleHealth)
	a.mux.HandleFunc("/visit", a.handleVisit)
}

func (a *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.mux.ServeHTTP(w, r)
}

func (a *App) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	count, err := a.readVisitCount()
	if err != nil {
		http.Error(w, "failed to read visit count", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, `<!doctype html>
<html>
<head>
  <meta charset="utf-8">
  <title>hello-deploy</title>
</head>
<body>
  <h1>%s</h1>
  Version 2
  <p>Visits: %d</p>
  <p>POST to /visit to increment the persisted counter.</p>
  <p>GET /health for health checks.</p>
</body>
</html>`, escapeHTML(a.message), count)
}

func (a *App) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	resp := map[string]string{
		"status":  "ok",
		"message": a.message,
	}
	_ = json.NewEncoder(w).Encode(resp)
}

func (a *App) handleVisit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	count, err := a.incrementVisitCount()
	if err != nil {
		http.Error(w, "failed to persist visit count", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	resp := map[string]any{
		"status": "ok",
		"visits": count,
	}
	_ = json.NewEncoder(w).Encode(resp)
}

func (a *App) visitsFile() string {
	return filepath.Join(a.dataDir, "visits.txt")
}

func (a *App) readVisitCount() (int, error) {
	if err := os.MkdirAll(a.dataDir, 0o755); err != nil {
		return 0, err
	}

	path := a.visitsFile()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			if err := os.WriteFile(path, []byte("0"), 0o644); err != nil {
				return 0, err
			}
			return 0, nil
		}
		return 0, err
	}

	value := strings.TrimSpace(string(data))
	if value == "" {
		return 0, nil
	}

	count, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("invalid visit count: %w", err)
	}

	return count, nil
}

func (a *App) incrementVisitCount() (int, error) {
	count, err := a.readVisitCount()
	if err != nil {
		return 0, err
	}

	count++

	if err := os.WriteFile(a.visitsFile(), []byte(strconv.Itoa(count)), 0o644); err != nil {
		return 0, err
	}

	return count, nil
}

func getenv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func escapeHTML(s string) string {
	replacer := strings.NewReplacer(
		"&", "&amp;",
		"<", "&lt;",
		">", "&gt;",
		`"`, "&quot;",
		"'", "&#39;",
	)
	return replacer.Replace(s)
}

func listenAddr(host, port string) string {
	if host == "" {
		return ":" + port
	}
	return net.JoinHostPort(host, port)
}

func main() {
	app := NewApp()
	host := getenv("HOST", "0.0.0.0")
	port := getenv("PORT", "8080")
	addr := listenAddr(host, port)

	log.Printf("starting server on %s", addr)
	log.Printf("data dir: %s", app.dataDir)

	if err := http.ListenAndServe(addr, app); err != nil {
		log.Fatal(err)
	}
}
