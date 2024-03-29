package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	pb "github.com/tank4gun/gourlshortener/internal/pkg/proto"

	"github.com/tank4gun/gourlshortener/internal/app/db"
	"github.com/tank4gun/gourlshortener/internal/app/handlers"
	"github.com/tank4gun/gourlshortener/internal/app/server"
	"github.com/tank4gun/gourlshortener/internal/app/storage"
	"github.com/tank4gun/gourlshortener/internal/app/types"
	"github.com/tank4gun/gourlshortener/internal/app/varprs"
	"google.golang.org/grpc"
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
	deleteChannel := make(chan types.RequestToDelete, 10)
	currentServer := server.CreateServer(strg, deleteChannel)

	sigChan := make(chan os.Signal, 1)
	serverStoppedChan := make(chan struct{})
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	listen, err := net.Listen("tcp", varprs.GRPCServerAddress)
	if err != nil {
		log.Fatal(err)
	}
	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(handlers.UserIDInterceptor))
	pb.RegisterShortenderServer(grpcServer, handlers.NewShortenderServer(strg, deleteChannel))
	go func() {
		<-sigChan
		close(deleteChannel)
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		if err := currentServer.Shutdown(ctx); err != nil {
			log.Fatalf("Err while Shutdown, %v", err)
		}
		grpcServer.GracefulStop()
		close(serverStoppedChan)
		defer cancel()
	}()

	go func() {
		if err := grpcServer.Serve(listen); err != nil {
			log.Fatal(err)
		}
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
