package server

import (
	"github.com/tank4gun/gourlshortener/internal/app/handlers"
	"github.com/tank4gun/gourlshortener/internal/app/storage"
	"net/http"
)

func CreateServer(startStorage *storage.Storage) *http.Server {
	mux := http.NewServeMux()
	handlerWithStorage := handlers.NewHandlerWithStorage(startStorage)
	mux.HandleFunc("/", handlerWithStorage.URLHandler)
	server := &http.Server{
		Addr:    "localhost:8080",
		Handler: mux,
	}
	return server
}
