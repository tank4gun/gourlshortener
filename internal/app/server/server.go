package server

import (
	"compress/gzip"
	"github.com/go-chi/chi/v5"
	"github.com/tank4gun/gourlshortener/internal/app/handlers"
	"github.com/tank4gun/gourlshortener/internal/app/storage"
	"github.com/tank4gun/gourlshortener/internal/app/variables_parsing"
	"io"
	"net/http"
	"strings"
)

func ReceiveCompressed(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		} else {
			uncompressed, err := gzip.NewReader(r.Body)
			if err != nil {
				io.WriteString(w, err.Error())
				return
			}
			defer uncompressed.Close()
			r.Body = uncompressed
			next.ServeHTTP(w, r)
		}
	})
}

type gzipWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

func (w gzipWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func SendCompressed(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		} else {
			compressed := gzip.NewWriter(w)
			defer compressed.Close()
			w.Header().Set("Content-Encoding", "gzip")
			next.ServeHTTP(gzipWriter{ResponseWriter: w, Writer: compressed}, r)
		}
	})
}

func CreateServer(startStorage *storage.Storage) *http.Server {
	router := chi.NewRouter()
	router.Use(ReceiveCompressed)
	router.Use(SendCompressed)
	handlerWithStorage := handlers.NewHandlerWithStorage(startStorage)
	router.Post("/", handlerWithStorage.CreateShortURLHandler)
	router.Get("/{id}", handlerWithStorage.GetURLByIDHandler)
	router.Post("/api/shorten", handlerWithStorage.CreateShortenURLFromBodyHandler)

	server := &http.Server{
		Addr:    variables_parsing.ServerAddress,
		Handler: router,
	}
	return server
}
