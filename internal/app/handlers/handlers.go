// Package handlers contains all handlers for URLShortener service.
package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"math"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/tank4gun/gourlshortener/internal/app/storage"
	"github.com/tank4gun/gourlshortener/internal/app/varprs"
)

// userCtxName - type string
type userCtxName string

// UserIDCtxName - context key for userID
var UserIDCtxName = userCtxName("UserID")

// CookieKey - key for cookie generator
var CookieKey = []byte("URL-Shortener-Key")

// URLShortenderCookieName - cookie name
var URLShortenderCookieName = "URL-Shortener"

// RequestToDelete - message type for URL deletion
type RequestToDelete struct {
	URLs   []string // URLs - list with URLs to delete
	UserID uint     // UserID - user ID for URLs to delete
}

// HandlerWithStorage is used for storing all info about URLShortener service objects and handling requests.
type HandlerWithStorage struct {
	// storage - storage.IRepository implementation
	storage storage.IRepository
	// baseURL - base URL for shorten URLs, i.e. http://localhost:8080
	baseURL string
	// deleteChannel - channel for RequestToDelete object to process
	deleteChannel chan RequestToDelete
}

// URLBodyRequest is a base structure for request
type URLBodyRequest struct {
	// URL to shorten
	URL string `json:"url"`
}

// ShortenURLResponse response for shorten URL creation
type ShortenURLResponse struct {
	// URL result
	URL string `json:"result"`
}

// BatchURLRequest request type for batch URLs
type BatchURLRequest struct {
	// CorrelationID - ID for original-shorten URL match
	CorrelationID string `json:"correlation_id"`
	// OriginalURL - URL to shorten
	OriginalURL string `json:"original_url"`
}

// BatchURLResponse response type for batch URLs
type BatchURLResponse struct {
	// CorrelationID - ID for original-shorten URL match
	CorrelationID string `json:"correlation_id"`
	// ShortURL - result shorten URL
	ShortURL string `json:"short_url"`
}

// NewHandlerWithStorage creates HandlerWithStorage object with given storage.
func NewHandlerWithStorage(storageVal storage.IRepository) *HandlerWithStorage {
	return &HandlerWithStorage{storage: storageVal, baseURL: varprs.BaseURL, deleteChannel: make(chan RequestToDelete, 10)}
}

// ConvertShortURLBatchToIDs converts shorten URLs to list with IDs
func ConvertShortURLBatchToIDs(shortURLBatch []string) []uint {
	var result = make([]uint, 0)
	for _, shortURL := range shortURLBatch {
		result = append(result, ConvertShortURLToID(shortURL))
	}
	return result
}

// ConvertShortURLToID converts shorten URLs to its ID
func ConvertShortURLToID(shortURL string) uint {
	var id uint = 0
	var charToIndex = make(map[int32]uint)
	for index, val := range storage.AllPossibleChars {
		charToIndex[val] = uint(index)
	}
	for index, value := range shortURL {
		id += charToIndex[value] * uint(math.Pow(62, float64(len(shortURL)-index-1)))
	}
	return id
}

// DeleteURLsDaemon runs daemon for urls deletion.
func (strg *HandlerWithStorage) DeleteURLsDaemon() {
	for reqToDelete := range strg.deleteChannel {
		log.Printf("Got request to delete %d", reqToDelete.UserID)
		URLIDs := ConvertShortURLBatchToIDs(reqToDelete.URLs)
		log.Printf("Got URLIDs %v", URLIDs)
		_ = strg.storage.MarkBatchAsDeleted(URLIDs, reqToDelete.UserID)
	}
	close(strg.deleteChannel)
}

// CreateShortURLByURL creates short URL by given URL and inserts it into storage.
func (strg *HandlerWithStorage) CreateShortURLByURL(url string, userID uint) (shortURLResult string, errMsg string, errCode int) {
	currInd, indErr := strg.storage.GetNextIndex()
	if indErr != nil {
		return "", "Bad next index", http.StatusInternalServerError
	}
	strgErr := strg.storage.InsertValue(url, userID)
	var exErr *storage.ExistError
	log.Println(strgErr)
	if errors.As(strgErr, &exErr) {
		return storage.CreateShortURL(exErr.ID), "", http.StatusConflict
	}
	if strgErr != nil {
		return "", strgErr.Error(), http.StatusInternalServerError
	}
	shortURL := storage.CreateShortURL(currInd)
	return shortURL, "", 0
}

// CreateShortURLBatch creates short URLs by given URLs batch and inserts them into storage.
func (strg *HandlerWithStorage) CreateShortURLBatch(batchURLs []BatchURLRequest, userID uint) ([]BatchURLResponse, string, int) {
	currInd, indErr := strg.storage.GetNextIndex()
	if indErr != nil {
		return make([]BatchURLResponse, 0), "Bad next index", http.StatusInternalServerError
	}
	var resultURLs []BatchURLResponse
	var insertURLs []string
	for index, URLrequest := range batchURLs {
		shortURL := storage.CreateShortURL(currInd + uint(index))
		insertURLs = append(insertURLs, URLrequest.OriginalURL)
		resultURL := BatchURLResponse{CorrelationID: URLrequest.CorrelationID, ShortURL: strg.baseURL + shortURL}
		resultURLs = append(resultURLs, resultURL)
	}
	err := strg.storage.InsertBatchValues(insertURLs, currInd, userID)
	if err != nil {
		return make([]BatchURLResponse, 0), "Error while inserting into storage", http.StatusInternalServerError
	}
	return resultURLs, "", 0
}

// GetURLByIDHandler returns full URL by its ID if it exists
func (strg *HandlerWithStorage) GetURLByIDHandler(w http.ResponseWriter, r *http.Request) {
	shortURL := chi.URLParam(r, "id")
	id := ConvertShortURLToID(shortURL)
	url, errCode := strg.storage.GetValueByKeyAndUserID(id, r.Context().Value(UserIDCtxName).(uint))
	if errCode != 0 {
		http.Error(w, "Couldn't find url for id "+shortURL, errCode)
		return
	}
	w.Header().Set("Location", url)
	w.WriteHeader(http.StatusTemporaryRedirect)
	var empty []byte
	w.Write(empty)
}

// CreateShortURLHandler converts URL from request body to shorten one and saves into db
func (strg *HandlerWithStorage) CreateShortURLHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	url, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Got bad body content", http.StatusBadRequest)
		return
	}
	shortURL, errorMessage, errorCode := strg.CreateShortURLByURL(string(url), r.Context().Value(UserIDCtxName).(uint))
	if errorCode != 0 && errorCode != http.StatusConflict {
		http.Error(w, errorMessage, errorCode)
		return
	}
	if errorCode == http.StatusConflict {
		w.WriteHeader(http.StatusConflict)
	} else {
		w.WriteHeader(http.StatusCreated)
	}
	_, errWrite := w.Write([]byte(strg.baseURL + shortURL))
	if errWrite != nil {
		http.Error(w, "Bad code", http.StatusInternalServerError)
	}
}

// CreateShortenURLFromBodyHandler converts URL from json object to shorten one and saves into db
func (strg *HandlerWithStorage) CreateShortenURLFromBodyHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	jsonBody, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var requestURL URLBodyRequest
	err = json.Unmarshal(jsonBody, &requestURL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if requestURL.URL == "" {
		http.Error(w, "Got empty url in Body", http.StatusUnprocessableEntity)
		return
	}
	shortURL, errorMessage, errorCode := strg.CreateShortURLByURL(requestURL.URL, r.Context().Value(UserIDCtxName).(uint))
	if errorCode != 0 && errorCode != http.StatusConflict {
		http.Error(w, errorMessage, errorCode)
		return
	}
	resultResponse := ShortenURLResponse{strg.baseURL + shortURL}
	w.Header().Set("Content-Type", "application/json")
	if errorCode == http.StatusConflict {
		w.WriteHeader(http.StatusConflict)
	} else {
		w.WriteHeader(http.StatusCreated)
	}
	if responseMarshalled, err := json.Marshal(resultResponse); err == nil {
		_, err = w.Write(responseMarshalled)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// CreateShortenURLBatchHandler converts URL batch from json object to shorten one and saves into db
func (strg *HandlerWithStorage) CreateShortenURLBatchHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	jsonBody, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var batchURLs []BatchURLRequest
	err = json.Unmarshal(jsonBody, &batchURLs)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	resultURLs, errorMessage, errorCode := strg.CreateShortURLBatch(batchURLs, r.Context().Value(UserIDCtxName).(uint))

	if errorCode != 0 {
		http.Error(w, errorMessage, errorCode)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if resultURLsMarshalled, err := json.Marshal(resultURLs); err == nil {
		_, err := w.Write(resultURLsMarshalled)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// GetAllURLsHandler return all URLs for given User
func (strg *HandlerWithStorage) GetAllURLsHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(UserIDCtxName).(uint)
	responseList, errCode := strg.storage.GetAllURLsByUserID(userID, strg.baseURL)
	if errCode != http.StatusOK {
		w.WriteHeader(errCode)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if responseMarshalled, err := json.Marshal(responseList); err == nil {
		_, err = w.Write(responseMarshalled)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// PingHandler checks than connection to storage is alive
func (strg *HandlerWithStorage) PingHandler(w http.ResponseWriter, r *http.Request) {
	err := strg.storage.Ping()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	var empty []byte
	w.Write(empty)
}

// DeleteURLsHandler removes all URLs for given User
func (strg *HandlerWithStorage) DeleteURLsHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(UserIDCtxName).(uint)
	defer r.Body.Close()
	jsonBody, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// URLsToDelete - list with URLs to delete
	var URLsToDelete []string
	err = json.Unmarshal(jsonBody, &URLsToDelete)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	go func() {
		strg.deleteChannel <- RequestToDelete{URLs: URLsToDelete, UserID: userID}
	}()
	w.WriteHeader(http.StatusAccepted)
	var empty []byte
	w.Write(empty)
}
