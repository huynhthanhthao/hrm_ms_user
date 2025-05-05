package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"user/ent"
	"user/ent/proto/entpb"
	"user/internal/handler"
	"user/internal/router"

	"github.com/joho/godotenv"
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
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")
	sslmode := os.Getenv("DB_SSLMODE")

	// Kiểm tra thiếu biến
	if host == "" || port == "" || user == "" || dbname == "" {
			log.Fatal("❌ One or more required DB environment variables are not set")
	}

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
			host, port, user, password, dbname, sslmode)

	client, err := ent.Open("postgres", dsn)
	if err != nil {
			log.Fatalf("❌ failed opening connection to postgres: %v", err)
	}

	log.Println("✅ Connected to PostgreSQL")
	return client
}

func init() {
	err := godotenv.Load(".env")
	if err != nil {
			log.Fatalf("Error loading .env file")
	}
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