package handlers

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"github.com/tank4gun/gourlshortener/internal/app/storage"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetURLByIDHandler(t *testing.T) {
	type want struct {
		code          int
		headerContent string
	}
	tests := []struct {
		name           string
		want           want
		currentStorage storage.Storage
		url            string
	}{
		{
			name: "short_url_exists",
			want: want{
				307,
				"http://ya.ru",
			},
			currentStorage: storage.Storage{InternalStorage: map[uint]string{1: "http://ya.ru"}, NextIndex: 2},
			url:            "/b",
		},
		{
			name: "short_url_does_not_exists",
			want: want{
				400,
				"",
			},
			currentStorage: storage.Storage{InternalStorage: map[uint]string{2: "http://ya.ru"}, NextIndex: 3},
			url:            "/b",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, tt.url, nil)
			w := httptest.NewRecorder()
			handler := http.HandlerFunc(NewHandlerWithStorage(&tt.currentStorage).GetURLByIDHandler)
			handler.ServeHTTP(w, request)
			result := w.Result()
			defer result.Body.Close()
			assert.Equal(t, tt.want.code, result.StatusCode)
			value := result.Header.Get("Location")
			assert.Equal(t, tt.want.headerContent, value)
		})
	}
}

func TestCreateShortURLHandler(t *testing.T) {
	type want struct {
		code            int
		responseContent string
	}
	tests := []struct {
		name            string
		want            want
		previousStorage storage.Storage
		resultStorage   storage.Storage
		url             string
	}{
		{
			name: "url_creation_success",
			want: want{
				201,
				"http://localhost:8080/b",
			},
			previousStorage: storage.Storage{InternalStorage: map[uint]string{}, NextIndex: 1},
			resultStorage:   storage.Storage{InternalStorage: map[uint]string{1: "http://ya.ru"}, NextIndex: 2},
			url:             "http://ya.ru",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte(tt.url)))
			w := httptest.NewRecorder()
			handler := http.HandlerFunc(NewHandlerWithStorage(&tt.previousStorage).CreateShortURLHandler)
			handler.ServeHTTP(w, request)
			result := w.Result()
			assert.Equal(t, tt.want.code, result.StatusCode)
			defer result.Body.Close()
			responseBody, err := io.ReadAll(result.Body)
			assert.Nil(t, err)
			assert.Equal(t, tt.want.responseContent, string(responseBody))
			assert.Equal(t, tt.previousStorage.InternalStorage, tt.resultStorage.InternalStorage)
			assert.Equal(t, tt.previousStorage.NextIndex, tt.resultStorage.NextIndex)
		})
	}
}

func TestConvertShortURLToID(t *testing.T) {
	tests := []struct {
		name       string
		shortURL   string
		expectedID uint
	}{
		{
			"b_to_1",
			"b",
			1,
		},
		{
			"c_to_2",
			"c",
			2,
		},
		{
			"cb_to_125",
			"cb",
			125,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ConvertShortURLToID(tt.shortURL)
			assert.Equal(t, tt.expectedID, result)
		})
	}
}

func TestCreateShortURL(t *testing.T) {
	tests := []struct {
		name             string
		index            uint
		expectedShortURL string
	}{
		{
			"1_to_b",
			1,
			"b",
		},
		{
			"2_to_c",
			2,
			"c",
		},
		{
			"125_to_bc",
			125,
			"bc",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CreateShortURL(tt.index)
			assert.Equal(t, tt.expectedShortURL, result)
		})
	}
}
