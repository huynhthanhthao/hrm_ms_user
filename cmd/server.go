package main

import (
	"context"
	"log"
	"net"
	"os"
	"user/ent"
	"user/ent/proto/entpb"
	"user/internal/handler"
	"user/internal/router"

	_ "github.com/lib/pq"
	"google.golang.org/grpc"
)

const (
	httpPort = ":8089"
	grpcPort = ":50051"
)

var (
	ctx    = context.Background()
	logger = log.New(os.Stdout, "[user-service] ", log.LstdFlags)
)

func main() {
	client := initEntClient()
	defer client.Close()

	runMigration(client)

	go startGRPCServer(client)
	startHTTPServer(client)
}

// Initialize Ent client
func initEntClient() *ent.Client {
	client, err := ent.Open("postgres", "host=localhost port=5432 user=postgres password=0000 dbname=user_service sslmode=disable")
	if err != nil {
		logger.Fatalf("❌ failed opening connection to postgres: %v", err)
	}
	logger.Println("✅ Connected to PostgreSQL")
	return client
}

// Run schema migration
func runMigration(client *ent.Client) {
	if err := client.Schema.Create(ctx); err != nil {
		logger.Fatalf("❌ failed creating schema resources: %v", err)
	}
	logger.Println("✅ Database schema created with Ent")
}

// Start gRPC server
func startGRPCServer(client *ent.Client) {
	grpcSvc := entpb.NewUserService(client)
	grpcServer := grpc.NewServer()
	entpb.RegisterUserServiceServer(grpcServer, grpcSvc)

	lis, err := net.Listen("tcp", grpcPort)
	if err != nil {
		logger.Fatalf("❌ failed to listen for gRPC: %v", err)
	}

	logger.Printf("✅ gRPC server listening on %s", grpcPort)
	if err := grpcServer.Serve(lis); err != nil {
		logger.Fatalf("❌ gRPC server stopped: %v", err)
	}
}

// Start HTTP server
func startHTTPServer(client *ent.Client) {
	r := router.SetupRouter(client)

	r.Use(handler.Logger())

	logger.Printf("✅ HTTP server listening on %s", httpPort)
	
	if err := r.Run(httpPort); err != nil {
		logger.Fatalf("❌ HTTP server stopped: %v", err)
	}
}