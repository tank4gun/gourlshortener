package app

import (
	"io"
	"math"
	"net/http"
	"strings"
)

var UrlMap = make(map[int]string)
var NextIndex = 1
var AllPossibleChars = "abcdefghijklmnopqrstuvwxwzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func ConvertUrlToShort(url string) string {
	var sb strings.Builder
	var currInd = NextIndex
	UrlMap[NextIndex] = url
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

func ConvertShortUrlToId(shortUrl string) int {
	id := 0
	var charToIndex = make(map[int32]int)
	for index, val := range AllPossibleChars {
		charToIndex[val] = index
	}
	for index, value := range shortUrl {
		id += charToIndex[value] * int(math.Pow(62, float64(len(shortUrl)-index-1)))
	}
	return id
}

func GetUrlByIdHandler(w http.ResponseWriter, r *http.Request) {
	shortUrl := r.URL.Query().Get("id")
	id := ConvertShortUrlToId(shortUrl)
	url, ok := UrlMap[id]
	if !ok {
		http.Error(w, "Couldn't find url for id "+shortUrl, 400)
		return
	}
	w.Header().Set("Location", url)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func CreateShortUrlHandler(w http.ResponseWriter, r *http.Request) {
	url, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Got bad body content", 400)
		return
	}
	shortUrl := ConvertUrlToShort(string(url))
	w.WriteHeader(201)
	w.Write([]byte(shortUrl))
}

func UrlHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		GetUrlByIdHandler(w, r)
	case http.MethodPost:
		CreateShortUrlHandler(w, r)
	default:
		http.Error(w, "Couldn't process request", 400)
		return
	}
}
