package main

import (
	"github.com/tank4gun/gourlshortener/internal/app/server"
	"github.com/tank4gun/gourlshortener/internal/app/storage"
	"log"
)

func main() {
	internalStorage := map[uint]string{}
	nextIndex := uint(1)
	strg := &storage.Storage{InternalStorage: internalStorage, NextIndex: nextIndex}
	currentServer := server.CreateServer(strg)
	log.Fatal(currentServer.ListenAndServe())
}
