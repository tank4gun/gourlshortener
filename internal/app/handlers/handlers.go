package handlers

import (
	"encoding/json"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/tank4gun/gourlshortener/internal/app/storage"
	"github.com/tank4gun/gourlshortener/internal/app/varprs"
	"io"
	"log"
	"math"
	"net/http"
)

type userCtxName string

var UserIDCtxName = userCtxName("UserID")
var CookieKey = []byte("URL-Shortener-Key")
var URLShortenderCookieName = "URL-Shortener"

type RequestToDelete struct {
	URLs   []string
	UserID uint
}

type HandlerWithStorage struct {
	storage storage.Repository
	//db      *sql.DB
	baseURL       string
	deleteChannel chan RequestToDelete
}

type URLBodyRequest struct {
	URL string `json:"url"`
}

type ShortenURLResponse struct {
	URL string `json:"result"`
}

type BatchURLRequest struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type BatchURLResponse struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

func NewHandlerWithStorage(storageVal storage.Repository) *HandlerWithStorage {
	return &HandlerWithStorage{storage: storageVal, baseURL: varprs.BaseURL, deleteChannel: make(chan RequestToDelete, 1)}
}

func ConvertShortURLBatchToIDs(shortURLBatch []string) []uint {
	var result = make([]uint, 0)
	for _, shortURL := range shortURLBatch {
		result = append(result, ConvertShortURLToID(shortURL))
	}
	return result
}

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

func (strg *HandlerWithStorage) DeleteURLsDaemon() {
	for reqToDelete := range strg.deleteChannel {
		log.Printf("Got request to delete %d", reqToDelete.UserID)
		URLIDs := ConvertShortURLBatchToIDs(reqToDelete.URLs)
		log.Printf("Got URLIDs %v", URLIDs)
		_ = strg.storage.MarkBatchAsDeleted(URLIDs, reqToDelete.UserID)
	}
	//close(strg.deleteChannel)
}

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
		return "", "Couldn't insert new value into storage", http.StatusInternalServerError
	}
	shortURL := storage.CreateShortURL(currInd)
	return shortURL, "", 0
}

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

func (strg *HandlerWithStorage) DeleteURLs(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(UserIDCtxName).(uint)
	defer r.Body.Close()
	jsonBody, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
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
