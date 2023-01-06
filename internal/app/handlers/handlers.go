package handlers

import (
	"database/sql"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"github.com/tank4gun/gourlshortener/internal/app/storage"
	"github.com/tank4gun/gourlshortener/internal/app/varprs"
	"io"
	"math"
	"net/http"
	"strings"
)

type userCtxName string

var AllPossibleChars = "abcdefghijklmnopqrstuvwxwzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
var UserIDCtxName = userCtxName("UserID")
var CookieKey = []byte("URL-Shortener-Key")
var URLShortenderCookieName = "URL-Shortener"

type HandlerWithStorage struct {
	storage *storage.Storage
	db      *sql.DB
	baseURL string
}

type URLBodyRequest struct {
	URL string `json:"url"`
}

type ShortenURLResponse struct {
	URL string `json:"result"`
}

type FullInfoURLResponse struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

func NewHandlerWithStorage(storageVal *storage.Storage, db *sql.DB) *HandlerWithStorage {
	return &HandlerWithStorage{storage: storageVal, db: db, baseURL: varprs.BaseURL}
}

func CreateShortURL(currInd uint) string {
	var sb strings.Builder
	for {
		if currInd == 0 {
			break
		}
		sb.WriteByte(AllPossibleChars[currInd%62])
		currInd = currInd / 62
	}
	return sb.String()
}

func ConvertShortURLToID(shortURL string) uint {
	var id uint = 0
	var charToIndex = make(map[int32]uint)
	for index, val := range AllPossibleChars {
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
	shortURL := CreateShortURL(currInd)
	return shortURL, "", 0
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

func (strg *HandlerWithStorage) GetAllURLs(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(UserIDCtxName).(uint)
	userURLs, ok := strg.storage.UserIDToURLID[userID]
	if !ok {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	responseList := make([]FullInfoURLResponse, 0)
	for _, URLID := range userURLs {
		shortURL := CreateShortURL(URLID)
		shortURL = strg.baseURL + shortURL
		originalURL, ok := strg.storage.InternalStorage[URLID]
		if !ok {
			http.Error(w, "Could not get URL from storage by ID", http.StatusInternalServerError)
			return
		}
		responseList = append(responseList, FullInfoURLResponse{ShortURL: shortURL, OriginalURL: originalURL})
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
	err := strg.db.Ping()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.WriteHeader(http.StatusOK)
	var empty []byte
	w.Write(empty)
}
