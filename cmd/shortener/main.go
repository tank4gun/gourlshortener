package main

import (
	"github.com/tank4gun/gourlshortener/internal/app/server"
	"github.com/tank4gun/gourlshortener/internal/app/storage"
	"log"
)

func main() {
	log.Fatal(server.CreateServer(&storage.Storage{InternalStorage: map[uint]string{}, NextIndex: uint(1)}).ListenAndServe())
}
