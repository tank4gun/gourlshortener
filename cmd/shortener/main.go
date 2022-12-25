package main

import (
	"github.com/tank4gun/gourlshortener/internal/app/server"
	"github.com/tank4gun/gourlshortener/internal/app/storage"
	"github.com/tank4gun/gourlshortener/internal/app/variables_parsing"
	"log"
)

func main() {
	variables_parsing.Init()
	internalStorage := map[uint]string{}
	nextIndex := uint(1)
	strg, _ := storage.NewStorage(internalStorage, nextIndex, variables_parsing.FileStoragePath)
	currentServer := server.CreateServer(strg)
	log.Fatal(currentServer.ListenAndServe())
}
