package server

import (
	"github.com/go-chi/chi/v5"
	"github.com/tank4gun/gourlshortener/internal/app/handlers"
	"github.com/tank4gun/gourlshortener/internal/app/storage"
	"net/http"
)

func CreateServer(startStorage *storage.Storage) *http.Server {
	router := chi.NewRouter()
	handlerWithStorage := handlers.NewHandlerWithStorage(startStorage)
	router.Post("/", handlerWithStorage.CreateShortURLHandler)
	router.Get("/{id}", handlerWithStorage.GetURLByIDHandler)
	router.Post("/api/shorten", handlerWithStorage.CreateShortenURLFromBodyHandler)
	server := &http.Server{
		Addr:    "localhost:8080",
		Handler: router,
	}
	return server
}
