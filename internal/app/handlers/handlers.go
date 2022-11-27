package handlers

import (
	"github.com/tank4gun/gourlshortener/internal/app/storage"
	"io"
	"math"
	"net/http"
	"strings"
)

var AllPossibleChars = "abcdefghijklmnopqrstuvwxwzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

type HandlerWithStorage struct {
	storage *storage.Storage
}

func NewHandlerWithStorage(storageVal *storage.Storage) *HandlerWithStorage {
	return &HandlerWithStorage{storage: storageVal}
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

func (strg *HandlerWithStorage) GetURLByIDHandler(w http.ResponseWriter, r *http.Request) {
	shortURL := r.URL.Path[1:]
	id := ConvertShortURLToID(shortURL)
	url, ok := strg.storage.InternalStorage[id]
	if !ok {
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
	currInd := strg.storage.NextIndex
	strg.storage.NextIndex++
	strg.storage.InternalStorage[currInd] = string(url)
	shortURL := CreateShortURL(currInd)
	w.WriteHeader(201)
	_, errWrite := w.Write([]byte("http://localhost:8080/" + shortURL))
	if errWrite != nil {
		http.Error(w, "Bad code", 400)
	}
}

func (strg *HandlerWithStorage) URLHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		strg.GetURLByIDHandler(w, r)
	case http.MethodPost:
		strg.CreateShortURLHandler(w, r)
	default:
		http.Error(w, "Couldn't process request", 400)
		return
	}
}
