package server

import (
	"github.com/go-chi/chi/v5"
	"github.com/tank4gun/gourlshortener/internal/app/handlers"
	"github.com/tank4gun/gourlshortener/internal/app/storage"
	"net/http"
	"os"
)

func CreateServer(startStorage *storage.Storage) *http.Server {
	router := chi.NewRouter()
	handlerWithStorage := handlers.NewHandlerWithStorage(startStorage)
	router.Post("/", handlerWithStorage.CreateShortURLHandler)
	router.Get("/{id}", handlerWithStorage.GetURLByIDHandler)
	router.Post("/api/shorten", handlerWithStorage.CreateShortenURLFromBodyHandler)
	serverAddrEnv := os.Getenv("SERVER_ADDRESS")
	if serverAddrEnv == "" {
		serverAddrEnv = "localhost:8080"
	}
	server := &http.Server{
		Addr:    serverAddrEnv,
		Handler: router,
	}
	return server
}
