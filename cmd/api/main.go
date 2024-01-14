package main

import (
	"fmt"
	logProto "github.com/tom-blog-app/blog-proto/log"
	"github.com/tom-blog-app/log-service/pkg/service"
	"google.golang.org/grpc/reflection"
	"os"
	"sync"

	//"os"

	db "github.com/tom-blog-app/blog-utils/database"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"net"

	"google.golang.org/grpc"

	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

var (
	mongoURL = os.Getenv("LOG_MONGO_URL")
	gRpcPort = os.Getenv("LOG_GRPC_PORT")
)

type LogServiceApp struct {
	server *grpc.Server
	client *mongo.Client
}

func NewLogServiceApp() *LogServiceApp {

	client, err := db.ConnectToMongo(mongoURL)

	if err != nil {
		log.Panic(err)
	}

	return &LogServiceApp{
		server: grpc.NewServer(),
		client: client,
	}
}

func main() {

	logApp := NewLogServiceApp()
	reflection.Register(logApp.server)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		logApp.registerService()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		logApp.checkHealth()
	}()

	wg.Wait()
}

func (app *LogServiceApp) registerService() {
	log.Println("Registering gRPC server..." + gRpcPort)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", gRpcPort))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	logProto.RegisterLogServiceServer(app.server, &service.LogServer{
		Client: app.client,
	})

	log.Printf("gRPC Server started on port %s", gRpcPort)

	if err := app.server.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func (app *LogServiceApp) checkHealth() {
	healthServer := health.NewServer()
	healthServer.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
	healthpb.RegisterHealthServer(app.server, healthServer)
}
