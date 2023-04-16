package handlers

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/tank4gun/gourlshortener/internal/app/storage"
)

func ExampleHandlerWithStorage_CreateShortenURLFromBodyHandler() {
	request := httptest.NewRequest(
		http.MethodPost, "/api/shorten", bytes.NewReader([]byte(`{"url": "http://ya.ru"}`)),
	)
	w := httptest.NewRecorder()
	ctx := context.WithValue(request.Context(), UserIDCtxName, uint(1))
	request = request.WithContext(ctx)
	handler := http.HandlerFunc(NewHandlerWithStorage(&storage.Storage{
		InternalStorage: map[uint]storage.URL{}, UserIDToURLID: map[uint][]uint{},
		NextIndex: 1, Encoder: nil, Decoder: nil,
	}, make(chan RequestToDelete, 10)).CreateShortenURLFromBodyHandler)
	handler.ServeHTTP(w, request)
	result := w.Result()
	fmt.Println(result.StatusCode)
	fmt.Println(result.Header.Get("Content-Type"))
	defer result.Body.Close()
	responseBody, _ := io.ReadAll(result.Body)
	fmt.Println(string(responseBody))

	// Output:
	// 201
	// application/json
	// {"result":"http://localhost:8080/b"}
}

func ExampleHandlerWithStorage_CreateShortURLHandler() {
	request := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte("http://ya.ru")))
	w := httptest.NewRecorder()
	ctx := context.WithValue(request.Context(), UserIDCtxName, uint(1))
	request = request.WithContext(ctx)
	handler := http.HandlerFunc(NewHandlerWithStorage(&storage.Storage{
		InternalStorage: map[uint]storage.URL{}, UserIDToURLID: map[uint][]uint{}, NextIndex: 1, Encoder: nil, Decoder: nil,
	}, make(chan RequestToDelete, 10)).CreateShortURLHandler)
	handler.ServeHTTP(w, request)
	result := w.Result()
	fmt.Println(result.StatusCode)
	defer result.Body.Close()
	responseBody, _ := io.ReadAll(result.Body)
	fmt.Println(string(responseBody))

	// Output:
	// 201
	// http://localhost:8080/b
}
