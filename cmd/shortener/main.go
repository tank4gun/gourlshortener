package main

import (
	"github.com/tank4gun/gourlshortener/internal/app/db"
	"github.com/tank4gun/gourlshortener/internal/app/server"
	"github.com/tank4gun/gourlshortener/internal/app/storage"
	"github.com/tank4gun/gourlshortener/internal/app/varprs"
	"log"
)

func main() {
	varprs.Init()
	db.RunMigrations(varprs.DatabaseDSN)
	internalStorage := map[uint]string{}
	nextIndex := uint(1)
	strg, _ := storage.NewStorage(internalStorage, nextIndex, varprs.FileStoragePath, varprs.DatabaseDSN)
	//database, _ := db.CreateDB(varprs.DatabaseDSN)
	currentServer := server.CreateServer(strg)
	log.Fatal(currentServer.ListenAndServe())
}
