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

func Init() {
	flag.StringVar(&server.ServerAddress, "a", "localhost:8080", "Server address")
	flag.StringVar(&handlers.BaseURL, "b", "http://localhost:8080", "Base URL for shorten URLs")
	flag.StringVar(&FileStoragePath, "f", "storage.txt", "File path for storage")
	flag.Parse()

	fileStoragePathEnv := os.Getenv("FILE_STORAGE_PATH")
	if fileStoragePathEnv != "" {
		FileStoragePath = fileStoragePathEnv
	}

	baseURLEnv := os.Getenv("BASE_URL")
	if baseURLEnv != "" {
		handlers.BaseURL = baseURLEnv
	} else {
		if handlers.BaseURL == "" {
			handlers.BaseURL = "http://localhost:8080"
		}
	}
	handlers.BaseURL += "/"

	serverAddrEnv := os.Getenv("SERVER_ADDRESS")
	if serverAddrEnv != "" {
		server.ServerAddress = serverAddrEnv
	} else {
		if server.ServerAddress == "" {
			server.ServerAddress = "localhost:8080"
		}
	}
}

func main() {
	Init()
	internalStorage := map[uint]string{}
	nextIndex := uint(1)
	strg, _ := storage.NewStorage(internalStorage, nextIndex, FileStoragePath)
	currentServer := server.CreateServer(strg)
	log.Fatal(currentServer.ListenAndServe())
}
