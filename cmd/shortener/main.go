package main

import (
	"github.com/tank4gun/gourlshortener/internal/app/server"
	"github.com/tank4gun/gourlshortener/internal/app/storage"
	"log"
	"os"
)

func main() {
	internalStorage := map[uint]string{}
	nextIndex := uint(1)
	strg, _ := storage.NewStorage(internalStorage, nextIndex, os.Getenv("FILE_STORAGE_PATH"))
	currentServer := server.CreateServer(strg)
	log.Fatal(currentServer.ListenAndServe())
}
