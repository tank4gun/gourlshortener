package server

import (
	"compress/gzip"
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/tank4gun/gourlshortener/internal/app/handlers"
	"github.com/tank4gun/gourlshortener/internal/app/storage"
	"github.com/tank4gun/gourlshortener/internal/app/varprs"
)

// GenerateNewID - generates new ID with 4 bytes for user randomly
func GenerateNewID() []byte {
	newData := make([]byte, 4)
	_, err := rand.Read(newData)
	if err != nil {
		panic(err.Error())
	}
	return newData
}

// ReceiveCompressed - middleware for uncompressing request body
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

// gzipWriter - struct for using as http.ResponseWriter in middleware
type gzipWriter struct {
	http.ResponseWriter
	// Writer - io.Writer
	Writer io.Writer
}

// Write - compress given bytes with gzip
func (w gzipWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

// SendCompressed - middleware for compressing response body
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

// CheckAuth - middleware for checking that user is authorized
func CheckAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := (*r).Cookie(handlers.URLShortenderCookieName)
		if cookie != nil && err != nil {
			fmt.Println(err.Error())
			io.WriteString(w, err.Error())
			return
		}
		if cookie != nil {
			cookieValue, _ := hex.DecodeString(cookie.Value)
			h := hmac.New(sha256.New, handlers.CookieKey)
			h.Write(cookieValue[:4])
			sign := h.Sum(nil)
			if hmac.Equal(sign, cookieValue[4:]) {
				ctx := context.WithValue(r.Context(), handlers.UserIDCtxName, uint(binary.BigEndian.Uint16(cookieValue[:4])))
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
		}

		newID := GenerateNewID()
		h := hmac.New(sha256.New, handlers.CookieKey)
		h.Write(newID)
		sign := h.Sum(nil)
		newCookie := http.Cookie{Name: handlers.URLShortenderCookieName, Value: hex.EncodeToString(append(newID[:], sign[:]...))}
		http.SetCookie(w, &newCookie)
		ctx := context.WithValue(r.Context(), handlers.UserIDCtxName, uint(binary.BigEndian.Uint16(newID)))
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// CreateServer - base method for creating Router and use it in http.Server
func CreateServer(startStorage storage.IRepository, deleteChannel chan handlers.RequestToDelete) *http.Server {
	router := chi.NewRouter()
	router.Use(ReceiveCompressed)
	router.Use(SendCompressed)
	router.Use(CheckAuth)
	handlerWithStorage := handlers.NewHandlerWithStorage(startStorage, deleteChannel)
	go handlerWithStorage.DeleteURLsDaemon()
	router.Post("/", handlerWithStorage.CreateShortURLHandler)
	router.Get("/{id}", handlerWithStorage.GetURLByIDHandler)
	router.Post("/api/shorten", handlerWithStorage.CreateShortenURLFromBodyHandler)
	router.Get("/api/user/urls", handlerWithStorage.GetAllURLsHandler)
	router.Delete("/api/user/urls", handlerWithStorage.DeleteURLsHandler)
	router.Get("/ping", handlerWithStorage.PingHandler)
	router.Post("/api/shorten/batch", handlerWithStorage.CreateShortenURLBatchHandler)
	router.Get("/api/internal/stats", handlerWithStorage.GetStatsHandler)

	// Add handlers for pprof
	router.Handle("/debug/pprof/*", http.DefaultServeMux)

	server := &http.Server{
		Addr:    varprs.ServerAddress,
		Handler: router,
	}
	return server
}
