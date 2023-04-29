// Package handlers contains all handlers for URLShortener service.
package handlers

import (
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"io"
	"log"
	"math"
	"net"
	"net/http"

	"github.com/tank4gun/gourlshortener/internal/app/storage"
	"github.com/tank4gun/gourlshortener/internal/app/types"
	"github.com/tank4gun/gourlshortener/internal/app/varprs"
)

// HandlerWithStorage is used for storing all info about URLShortener service objects and handling requests.
type HandlerWithStorage struct {
	// storage - storage.IRepository implementation
	storage storage.IRepository
	// baseURL - base URL for shorten URLs, i.e. http://localhost:8080
	baseURL string
	// deleteChannel - channel for RequestToDelete object to process
	deleteChannel chan types.RequestToDelete
}

// NewHandlerWithStorage creates HandlerWithStorage object with given storage.
func NewHandlerWithStorage(storageVal storage.IRepository, deleteChannel chan types.RequestToDelete) *HandlerWithStorage {
	return &HandlerWithStorage{storage: storageVal, baseURL: varprs.BaseURL, deleteChannel: deleteChannel}
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
}

//
//// CreateShortURLBatch creates short URLs by given URLs batch and inserts them into storage.
//func (strg *HandlerWithStorage) CreateShortURLBatch(batchURLs []storage.BatchURLRequest, userID uint) ([]storage.BatchURLResponse, string, int) {
//	currInd, indErr := strg.storage.GetNextIndex()
//	if indErr != nil {
//		return make([]storage.BatchURLResponse, 0), "Bad next index", http.StatusInternalServerError
//	}
//	var resultURLs []storage.BatchURLResponse
//	var insertURLs []string
//	for index, URLrequest := range batchURLs {
//		shortURL := storage.CreateShortURL(currInd + uint(index))
//		insertURLs = append(insertURLs, URLrequest.OriginalURL)
//		resultURL := storage.BatchURLResponse{CorrelationID: URLrequest.CorrelationID, ShortURL: strg.baseURL + shortURL}
//		resultURLs = append(resultURLs, resultURL)
//	}
//	err := strg.storage.InsertBatchValues(insertURLs, currInd, userID)
//	if err != nil {
//		return make([]storage.BatchURLResponse, 0), "Error while inserting into storage", http.StatusInternalServerError
//	}
//	return resultURLs, "", 0
//}

// GetURLByIDHandler returns full URL by its ID if it exists
func (strg *HandlerWithStorage) GetURLByIDHandler(w http.ResponseWriter, r *http.Request) {
	shortURL := chi.URLParam(r, "id")
	originalURL, errorCode := CommonServer{}.GetURLByID(strg.storage, shortURL, r.Context().Value(types.UserIDCtxName).(uint))
	if errorCode != 0 {
		http.Error(w, "Couldn't find url for id "+shortURL, errorCode)
		return
	}
	w.Header().Set("Location", originalURL)
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
	shortURL, errorMessage, errorCode := CommonServer{}.CreateShortURL(strg.storage, string(url), r.Context().Value(types.UserIDCtxName).(uint), strg.baseURL)
	if errorCode != 0 && errorCode != http.StatusConflict {
		http.Error(w, errorMessage, errorCode)
		return
	}
	if errorCode == http.StatusConflict {
		w.WriteHeader(http.StatusConflict)
	} else {
		w.WriteHeader(http.StatusCreated)
	}
	_, errWrite := w.Write([]byte(shortURL))
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
	var requestURL types.URLBodyRequest
	err = json.Unmarshal(jsonBody, &requestURL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if requestURL.URL == "" {
		http.Error(w, "Got empty url in Body", http.StatusUnprocessableEntity)
		return
	}
	shortURL, errorMessage, errorCode := strg.storage.CreateShortURLByURL(requestURL.URL, r.Context().Value(types.UserIDCtxName).(uint))
	if errorCode != 0 && errorCode != http.StatusConflict {
		http.Error(w, errorMessage, errorCode)
		return
	}
	resultResponse := types.ShortenURLResponse{URL: strg.baseURL + shortURL}
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
	var batchURLs []storage.BatchURLRequest
	err = json.Unmarshal(jsonBody, &batchURLs)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	resultURLs, errorMessage, errorCode := CommonServer{}.CreateShortenURLBatch(strg.storage, batchURLs, r.Context().Value(types.UserIDCtxName).(uint), strg.baseURL)
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
	userID := r.Context().Value(types.UserIDCtxName).(uint)
	responseList, errorCode := CommonServer{}.GetAllURLs(strg.storage, userID, strg.baseURL)
	if errorCode != http.StatusOK {
		w.WriteHeader(errorCode)
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

// DeleteURLsHandler removes all URLs for given User
func (strg *HandlerWithStorage) DeleteURLsHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(types.UserIDCtxName).(uint)
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
	CommonServer{}.DeleteURLs(strg.deleteChannel, URLsToDelete, userID)
	w.WriteHeader(http.StatusAccepted)
	var empty []byte
	w.Write(empty)
}

// PingHandler checks than connection to storage is alive
func (strg *HandlerWithStorage) PingHandler(w http.ResponseWriter, r *http.Request) {
	err := CommonServer{}.Ping(strg.storage)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	var empty []byte
	w.Write(empty)
}

// GetStatsHandler return all URLs and Users number
func (strg *HandlerWithStorage) GetStatsHandler(w http.ResponseWriter, r *http.Request) {
	ipStr := r.Header.Get("X-Real-IP")
	requestIP := net.ParseIP(ipStr)
	if requestIP == nil {
		http.Error(w, "Got bad IP address", http.StatusForbidden)
		return
	}
	_, ipNet, err := net.ParseCIDR(varprs.TrustedSubnet)
	if err != nil {
		http.Error(w, "Couldn't parse ipMask", http.StatusInternalServerError)
		return
	}
	if !ipNet.Contains(requestIP) {
		http.Error(w, "Got bad IP address", http.StatusForbidden)
		return
	}
	stats, errCode := CommonServer{}.GetStats(strg.storage)
	if errCode != http.StatusOK {
		w.WriteHeader(errCode)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if statsMarshalled, err := json.Marshal(stats); err == nil {
		_, err = w.Write(statsMarshalled)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
