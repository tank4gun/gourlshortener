package main

import (
	"flag"
	"github.com/tank4gun/gourlshortener/internal/app/handlers"
	"github.com/tank4gun/gourlshortener/internal/app/server"
	"github.com/tank4gun/gourlshortener/internal/app/storage"
	"log"
	"os"
)

var FileStoragePath string

func init() {
	flag.StringVar(&server.ServerAddress, "a", "localhost:8080", "Server address")
	flag.StringVar(&handlers.BaseURL, "b", "http://localhost:8080", "Base URL for shorten URLs")
	flag.StringVar(&FileStoragePath, "f", "storage.txt", "File path for storage")
}

func main() {
	flag.Parse()
	internalStorage := map[uint]string{}
	nextIndex := uint(1)
	fileStoragePath := os.Getenv("FILE_STORAGE_PATH")
	if fileStoragePath == "" {
		fileStoragePath = FileStoragePath
	}
	strg, _ := storage.NewStorage(internalStorage, nextIndex, fileStoragePath)
	currentServer := server.CreateServer(strg)
	log.Fatal(currentServer.ListenAndServe())
}
