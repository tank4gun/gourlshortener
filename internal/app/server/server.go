package server

import (
	"github.com/go-chi/chi/v5"
	"github.com/tank4gun/gourlshortener/internal/app/handlers"
	"github.com/tank4gun/gourlshortener/internal/app/storage"
	"net/http"
	"os"
)

var ServerAddress string

func CreateServer(startStorage *storage.Storage) *http.Server {
	router := chi.NewRouter()
	handlerWithStorage := handlers.NewHandlerWithStorage(startStorage)
	router.Post("/", handlerWithStorage.CreateShortURLHandler)
	router.Get("/{id}", handlerWithStorage.GetURLByIDHandler)
	router.Post("/api/shorten", handlerWithStorage.CreateShortenURLFromBodyHandler)
	serverAddr := os.Getenv("SERVER_ADDRESS")
	if serverAddr == "" {
		if ServerAddress == "" {
			serverAddr = "localhost:8080"
		} else {
			serverAddr = ServerAddress
		}
	}
	server := &http.Server{
		Addr:    serverAddr,
		Handler: router,
	}
	return server
}
