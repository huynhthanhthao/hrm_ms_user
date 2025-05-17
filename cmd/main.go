package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/huynhthanhthao/hrm_user_service/ent"
	"github.com/huynhthanhthao/hrm_user_service/ent/migrate"
	clientGrpc "github.com/huynhthanhthao/hrm_user_service/generated"

	userpb "github.com/huynhthanhthao/hrm_user_service/generated"
	userGrpc "github.com/huynhthanhthao/hrm_user_service/internal/grpc"
	"github.com/huynhthanhthao/hrm_user_service/internal/handler"
	"github.com/huynhthanhthao/hrm_user_service/internal/router"
	"github.com/huynhthanhthao/hrm_user_service/internal/service"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
)

var (
	logger   = log.New(os.Stdout, "[user-service] ", log.LstdFlags)
	httpPort = os.Getenv("HTTP_PORT")
	grpcPort = os.Getenv("GRPC_PORT")
)

func main() {
	client := initEntClient()
	defer client.Close()

	runMigration(client)

	hrClients, err := NewHRServiceClients()

	if err != nil {
		log.Fatalf("failed to initialize HR service clients: %v", err)
	}
	defer hrClients.Close()

	permissionClients, err := NewPermissionServiceClients()

	if err != nil {
		log.Fatalf("failed to initialize Permission service clients: %v", err)
	}
	defer permissionClients.Close()

	userService, err := service.NewUserService(client, hrClients, permissionClients)

	if err != nil {
		log.Fatalf("failed to initialize UserService: %v", err)
	}

	go startGRPCServer(client, userService)
	startHTTPServer(client, hrClients, permissionClients)
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
		log.Fatal("One or more required DB environment variables are not set")
	}

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode)

	client, err := ent.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("failed opening connection to postgres: %v", err)
	}

	log.Println("Connected to PostgreSQL")
	return client
}

func init() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	// Log loaded environment variables for debugging
	fmt.Printf("Loaded HTTP_PORT: %s", os.Getenv("HTTP_PORT"))
	fmt.Printf("Loaded GRPC_PORT: %s", os.Getenv("GRPC_PORT"))
	httpPort = ":" + os.Getenv("HTTP_PORT")
	grpcPort = ":" + os.Getenv("GRPC_PORT")
}

// Run schema migration
func runMigration(client *ent.Client) {
	ctx := context.Background()
	if err := client.Schema.Create(ctx, migrate.WithDropColumn(true)); err != nil {
		log.Fatalf("failed creating schema resources: %v", err)
	}
}

// Start gRPC server
func startGRPCServer(client *ent.Client, userService *service.UserService) {
	grpcServer := grpc.NewServer()

	userGrpcServer := userGrpc.NewUserGRPCServer(userService)
	userpb.RegisterUserServiceServer(grpcServer, userGrpcServer)

	reflection.Register(grpcServer)

	lis, err := net.Listen("tcp", grpcPort)
	if err != nil {
		logger.Fatalf("failed to listen for gRPC: %v", err)
	}

	logger.Printf("gRPC server listening on %s", grpcPort)

	if err := grpcServer.Serve(lis); err != nil {
		logger.Fatalf("gRPC server stopped: %v", err)
	}
}

// Start HTTP server
func startHTTPServer(client *ent.Client, hrClients *service.HRServiceClients, perClients *service.PermissionServiceClients,) {
	r := router.SetupRouter(client, hrClients, perClients)

	r.Use(handler.Logger())

	logger.Printf("HTTP server listening on %s", httpPort)

	if err := r.Run(httpPort); err != nil {
		logger.Fatalf("HTTP server stopped: %v", err)
	}
}

func NewHRServiceClients() (*service.HRServiceClients, error) {
	url := os.Getenv("HR_SERVICE_URL")
	if url == "" {
		return nil, fmt.Errorf("HR_SERVICE_URL is not set")
	}

	conn, err := grpc.NewClient(url, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to HR service: %v", err)
	}

	return &service.HRServiceClients{
		Conn:         conn,
		Organization: clientGrpc.NewOrganizationServiceClient(conn),
		HrExt:        clientGrpc.NewExtServiceClient(conn),
	}, nil
}

func NewPermissionServiceClients() (*service.PermissionServiceClients, error) {
	url := os.Getenv("PERMISSION_SERVICE_URL")
	if url == "" {
		return nil, fmt.Errorf("PERMISSION_SERVICE_URL is not set")
	}

	conn, err := grpc.NewClient(url, grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		return nil, fmt.Errorf("failed to connect to Permission service: %v", err)
	}

	return &service.PermissionServiceClients{
		Conn:     conn,
		UserRole: clientGrpc.NewUserRoleServiceClient(conn),
		UserPerm: clientGrpc.NewUserPermServiceClient(conn),
	}, nil
}
