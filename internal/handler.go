package internal

import (
	"net/http"

	"github.com/sebasromero/tfs-api/internal/files"
	"github.com/sebasromero/tfs-api/internal/health"
)

func MainHandler() *http.ServeMux {
	rootHandler := http.NewServeMux()

	rootHandler.HandleFunc("GET /health", health.Health)
	rootHandler.HandleFunc("GET /pull/{id}", files.Pull)
	rootHandler.HandleFunc("POST /push", files.Push)

	rootHandler.Handle("/api/v1/", http.StripPrefix("/api/v1", rootHandler))
	return rootHandler
}
