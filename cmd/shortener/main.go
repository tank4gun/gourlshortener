package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"

	"github.com/tank4gun/gourlshortener/internal/app/db"
	"github.com/tank4gun/gourlshortener/internal/app/server"
	"github.com/tank4gun/gourlshortener/internal/app/storage"
	"github.com/tank4gun/gourlshortener/internal/app/varprs"
)

// Use command `go run -ldflags "-X main.buildVersion=1.1.1 -X 'main.buildDate=$(date +'%Y/%m/%d %H:%M:%S')' -X main.buildCommit=123" shortener/main.go`
var buildVersion string
var buildDate string
var buildCommit string

func main() {
	if buildVersion == "" {
		buildVersion = "N/A"
	}
	if buildDate == "" {
		buildDate = "N/A"
	}
	if buildCommit == "" {
		buildCommit = "N/A"
	}
	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)

	varprs.Init()
	db.RunMigrations(varprs.DatabaseDSN)
	internalStorage := map[uint]storage.URL{}
	nextIndex := uint(1)
	strg, _ := storage.NewStorage(internalStorage, nextIndex, varprs.FileStoragePath, varprs.DatabaseDSN)
	currentServer := server.CreateServer(strg)

	sigChan := make(chan os.Signal, 1)
	serverStoppedChan := make(chan struct{})
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	go func() {
		<-sigChan
		if err := currentServer.Shutdown(context.Background()); err != nil {
			log.Fatalf("Err while Shutdown, %v", err)
		}
		close(serverStoppedChan)
	}()

	if varprs.UseHTTPS {
		if err := currentServer.ListenAndServeTLS("internal/app/varprs/localhost.crt", "internal/app/varprs/localhost.key"); err != nil {
			log.Fatalf("Err while ListenAndServeTLS, %v", err)
		}
	} else {
		if err := currentServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Err while ListenAndServe, %v", err)
		}
	}
	<-serverStoppedChan
	if err := strg.Shutdown(); err != nil {
		log.Fatalf("Err while Storage Shutdown, %v", err)
	}
	log.Println("Server was shutdowned")
}
