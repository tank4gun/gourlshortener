package app

import (
	"io"
	"math"
	"net/http"
	"strings"
)

var URLMap = make(map[int]string)
var NextIndex = 1
var AllPossibleChars = "abcdefghijklmnopqrstuvwxwzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func ConvertURLToShort(url string) string {
	var sb strings.Builder
	var currInd = NextIndex
	URLMap[NextIndex] = url
	for {
		if currInd == 0 {
			break
		}
		sb.WriteByte(AllPossibleChars[currInd%62])
		currInd = currInd / 62
	}
	NextIndex++
	return sb.String()
}

func ConvertShortURLToID(shortURL string) int {
	id := 0
	var charToIndex = make(map[int32]int)
	for index, val := range AllPossibleChars {
		charToIndex[val] = index
	}
	for index, value := range shortURL {
		id += charToIndex[value] * int(math.Pow(62, float64(len(shortURL)-index-1)))
	}
	return id
}

func GetURLByIDHandler(w http.ResponseWriter, r *http.Request) {
	shortURL := r.URL.Path[1:]
	id := ConvertShortURLToID(shortURL)
	url, ok := URLMap[id]
	if !ok {
		http.Error(w, "Couldn't find url for id "+shortURL, 400)
		return
	}
	w.Header().Set("Location", url)
	w.WriteHeader(http.StatusTemporaryRedirect)
	var empty []byte
	w.Write(empty)
}

func CreateShortURLHandler(w http.ResponseWriter, r *http.Request) {
	url, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Got bad body content", 400)
		return
	}
	shortURL := ConvertURLToShort(string(url))
	w.WriteHeader(201)
	_, errWrite := w.Write([]byte("http://localhost:8080/" + shortURL))
	if errWrite != nil {
		http.Error(w, "Bad code", 400)
	}
}

func URLHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		GetURLByIDHandler(w, r)
	case http.MethodPost:
		CreateShortURLHandler(w, r)
	default:
		http.Error(w, "Couldn't process request", 400)
		return
	}
}
