package server

import (
	"embed"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"

	"github.com/mr-karan/prom2grafana/internal/handlers"
)

// Server represents the HTTP server
type Server struct {
	port           string
	content        embed.FS
	convertHandler *handlers.ConvertHandler
}

// New creates a new server instance
func New(port string, content embed.FS, convertHandler *handlers.ConvertHandler) *Server {
	return &Server{
		port:           port,
		content:        content,
		convertHandler: convertHandler,
	}
}

// Start starts the HTTP server
func (s *Server) Start() error {
	// Create a sub filesystem for static files
	staticFS, err := fs.Sub(s.content, "static")
	if err != nil {
		return fmt.Errorf("failed to create static filesystem: %w", err)
	}

	// Serve static files from embedded filesystem
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))

	// Serve index.html from embedded filesystem
	http.HandleFunc("/", s.handleIndex)

	// API endpoint
	http.HandleFunc("/convert", s.convertHandler.Handle)

	slog.Info("Server starting", "port", s.port, "url", fmt.Sprintf("http://localhost:%s", s.port))
	return http.ListenAndServe(":"+s.port, nil)
}

// handleIndex serves the index.html file
func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	
	data, err := s.content.ReadFile("index.html")
	if err != nil {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if _, err := w.Write(data); err != nil {
		slog.Error("Failed to write response", "error", err)
	}
}