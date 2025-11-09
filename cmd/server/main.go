package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"datagateway/internal/db"
	"datagateway/internal/service"
	"datagateway/proto/userpb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	// 讀取 gRPC 監聽 port，預設 :50051
	addr := ":50051"
	if p := os.Getenv("GRPC_PORT"); p != "" {
		addr = ":" + p
	}

	// 初始化 DB
	gormDB, err := db.NewGormDB()
	if err != nil {
		log.Fatalf("failed to connect db: %v", err)
	}

	// 建立 gRPC server
	grpcServer := grpc.NewServer()
	userService := service.NewUserServiceServer(gormDB)
	userpb.RegisterUserServiceServer(grpcServer, userService)

	reflection.Register(grpcServer)

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// graceful shutdown
	go func() {
		log.Printf("gRPC server listening on %s", addr)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("gRPC serve error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	log.Println("shutting down gRPC server...")
	grpcServer.GracefulStop()
	// 關閉 DB 連線
	sqlDB, _ := gormDB.DB()
	_ = sqlDB.Close()
}
