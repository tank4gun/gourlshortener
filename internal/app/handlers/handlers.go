package handlers

import (
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"github.com/tank4gun/gourlshortener/internal/app/storage"
	"github.com/tank4gun/gourlshortener/internal/app/varprs"
	"io"
	"math"
	"net/http"
)

type userCtxName string

var UserIDCtxName = userCtxName("UserID")
var CookieKey = []byte("URL-Shortener-Key")
var URLShortenderCookieName = "URL-Shortener"

type HandlerWithStorage struct {
	storage storage.Repository
	//db      *sql.DB
	baseURL string
}

type URLBodyRequest struct {
	URL string `json:"url"`
}

type ShortenURLResponse struct {
	URL string `json:"result"`
}

type BatchUrlRequest struct {
	CorrelationId string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type BatchUrlResponse struct {
	CorrelationId string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

func NewHandlerWithStorage(storageVal storage.Repository) *HandlerWithStorage {
	return &HandlerWithStorage{storage: storageVal, baseURL: varprs.BaseURL}
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

func (strg *HandlerWithStorage) CreateShortURLByURL(url string, userID uint) (shortURLResult string, errMsg string, errCode int) {
	currInd, indErr := strg.storage.GetNextIndex()
	if indErr != nil {
		return "", "Bad next index", 500
	}
	strgErr := strg.storage.InsertValue(url, userID)
	if strgErr != nil {
		return "", "Couldn't insert new value into storage", 500
	}
	shortURL := storage.CreateShortURL(currInd)
	return shortURL, "", 0
}

func (strg *HandlerWithStorage) CreateShortURLBatch(batchURLs []BatchUrlRequest, userID uint) ([]BatchUrlResponse, string, int) {
	currInd, indErr := strg.storage.GetNextIndex()
	if indErr != nil {
		return make([]BatchUrlResponse, 0), "Bad next index", 500
	}
	var resultURLs []BatchUrlResponse
	var insertURLs []string
	for index, URLrequest := range batchURLs {
		shortURL := storage.CreateShortURL(currInd + uint(index))
		insertURLs = append(insertURLs, URLrequest.OriginalURL)
		resultURL := BatchUrlResponse{CorrelationId: URLrequest.CorrelationId, ShortURL: shortURL}
		resultURLs = append(resultURLs, resultURL)
	}
	err := strg.storage.InsertBatchValues(insertURLs, currInd, userID)
	if err != nil {
		return make([]BatchUrlResponse, 0), "Error while inserting into storage", 500
	}
	return resultURLs, "", 0
}

func (strg *HandlerWithStorage) GetURLByIDHandler(w http.ResponseWriter, r *http.Request) {
	shortURL := chi.URLParam(r, "id")
	id := ConvertShortURLToID(shortURL)
	url, err := strg.storage.GetValueByKeyAndUserID(id, r.Context().Value(UserIDCtxName).(uint))
	if err != nil {
		http.Error(w, "Couldn't find url for id "+shortURL, 400)
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
		http.Error(w, "Got bad body content", 400)
		return
	}
	shortURL, errorMessage, errorCode := strg.CreateShortURLByURL(string(url), r.Context().Value(UserIDCtxName).(uint))
	if errorCode != 0 {
		http.Error(w, errorMessage, errorCode)
		return
	}
	w.WriteHeader(201)
	_, errWrite := w.Write([]byte(strg.baseURL + shortURL))
	if errWrite != nil {
		http.Error(w, "Bad code", 500)
	}
}

func (strg *HandlerWithStorage) CreateShortenURLFromBodyHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	jsonBody, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	var requestURL URLBodyRequest
	err = json.Unmarshal(jsonBody, &requestURL)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	if requestURL.URL == "" {
		http.Error(w, "Got empty url in Body", http.StatusUnprocessableEntity)
		return
	}
	shortURL, errorMessage, errorCode := strg.CreateShortURLByURL(requestURL.URL, r.Context().Value(UserIDCtxName).(uint))
	if errorCode != 0 {
		http.Error(w, errorMessage, errorCode)
		return
	}
	resultResponse := ShortenURLResponse{strg.baseURL + shortURL}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(201)
	if responseMarshalled, err := json.Marshal(resultResponse); err == nil {
		_, err = w.Write(responseMarshalled)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
	} else {
		http.Error(w, err.Error(), 500)
		return
	}
}

func (strg *HandlerWithStorage) CreateShortenURLBatchHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	jsonBody, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	var batchURLs []BatchUrlRequest
	err = json.Unmarshal(jsonBody, &batchURLs)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	resultURLs, errorMessage, errorCode := strg.CreateShortURLBatch(batchURLs, r.Context().Value(UserIDCtxName).(uint))

	if errorCode != 0 {
		http.Error(w, errorMessage, errorCode)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	if resultURLsMarshalled, err := json.Marshal(resultURLs); err == nil {
		_, err := w.Write(resultURLsMarshalled)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
	} else {
		http.Error(w, err.Error(), 500)
		return
	}
}

func (strg *HandlerWithStorage) GetAllURLs(w http.ResponseWriter, r *http.Request) {
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
			http.Error(w, err.Error(), 500)
			return
		}
	} else {
		http.Error(w, err.Error(), 500)
		return
	}
}

func (strg *HandlerWithStorage) Ping(w http.ResponseWriter, r *http.Request) {
	err := strg.storage.Ping()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.WriteHeader(http.StatusOK)
	var empty []byte
	w.Write(empty)
}
