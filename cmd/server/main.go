package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"

	hellopb "Practice-gRPC/pkg/grpc"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type myServer struct {
	hellopb.UnimplementedGreetingServiceServer
}

func newMyServer() *myServer {
	return &myServer{}
}

func (s *myServer) Hello(ctx context.Context, req *hellopb.HelloRequest) (*hellopb.HelloResponse, error) {
	return &hellopb.HelloResponse{
		Message: fmt.Sprintf("Hello, %s!", req.GetName()),
	}, nil
}

func main() {
	port := 8080
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		panic(err)
	}

	s := grpc.NewServer()

	// サービスを登録
	hellopb.RegisterGreetingServiceServer(s, newMyServer())

	// gRPCurlにprotoファイルの情報を与えるため，サーバーリフレクションの設定
	reflection.Register(s)

	go func() {
		log.Printf("start gRPC server port: %v", port)
		s.Serve(listener)
	}()

	quit := make(chan os.Signal, 1)
	// Ctrl + C（SIGINT：割り込みシグナル）を受信するようチャネルquitを割り当て
	signal.Notify(quit, os.Interrupt)
	// チャネルquitがos.Signal型の値を受信すると，以降の処理が実行される
	<-quit
	log.Println("stopping gRPC server...")
	s.GracefulStop()
}
