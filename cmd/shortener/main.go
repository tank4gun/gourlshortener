package main

import (
	"github.com/tank4gun/gourlshortener/internal/app"
	"log"
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", app.UrlHandler)
	server := &http.Server{
		Addr:    "localhost:8080",
		Handler: mux,
	}
	log.Fatal(server.ListenAndServe())
}
