// cmd/server/main.go
package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/ghSurf/internal/config"
	"github.com/ghSurf/internal/githubclient"
	"github.com/ghSurf/internal/grpcserver"
	pb "github.com/ghSurf/proto"
)

func main() {
	// --- 1. Load Configuration ---
	// Specify the path to your .env file here
	envFilePath := "/Users/an_chou/envFiles/.env"

	cfg, err := config.Load(envFilePath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// --- 2. Setup Logger ---
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Printf("Log Level set to: %s", cfg.LogLevel)
	log.Printf("Port set to: %s", cfg.Port)

	// --- 3. Initialize Dependencies ---
	ghClient := githubclient.New(cfg.GithubToken)
	log.Println("GitHub client initialized.")

	ghSurfServer := grpcserver.NewGrpcServer(ghClient.Search)
	log.Println("gRPC server implementation initialized.")

	// --- 4. Setup TCP Listener ---
	listenAddr := fmt.Sprintf(":%s", cfg.Port)
	listener, err := net.Listen("tcp", listenAddr)
	if err != nil {
		log.Fatalf("FATAL: Failed to listen on %s: %v", listenAddr, err)
	}
	log.Printf("gRPC server listening on %s...", listenAddr)

	// --- 5. Create and Register gRPC Server ---
	grpcServer := grpc.NewServer()
	pb.RegisterGithubSearchServiceServer(grpcServer, ghSurfServer)
	log.Println("Registered GithubSearchService with gRPC server.")
	reflection.Register(grpcServer)
	log.Println("gRPC reflection service registered (for debugging).")

	// --- 6. Start gRPC Server Goroutine ---
	serveErrChan := make(chan error, 1)
	go func() {
		log.Println("Starting gRPC server...")
		serveErrChan <- grpcServer.Serve(listener)
	}()
	log.Println("gRPC server Serve() started in a goroutine.")

	// --- 7. Wait for Shutdown Signal ---
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, syscall.SIGINT, syscall.SIGTERM)
	log.Println("Waiting for shutdown signal (SIGINT/SIGTERM)...")

	select {
	case sig := <-stopChan:
		log.Printf("Received signal: %v. Initiating graceful shutdown...", sig)
	case err := <-serveErrChan:
		log.Printf("ERROR: gRPC server stopped unexpectedly: %v. Shutting down.", err)
	}

	// --- 8. Perform Graceful Shutdown ---
	log.Println("Calling gRPC server GracefulStop()...")
	grpcServer.GracefulStop()
	log.Println("ghSurf gRPC server stopped gracefully.")
}
