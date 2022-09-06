package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"time"

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

func (s *myServer) HelloServerStream(req *hellopb.HelloRequest, stream hellopb.GreetingService_HelloServerStreamServer) error {
	resCount := 5
	for i := 0; i < resCount; i++ {
		if err := stream.Send(&hellopb.HelloResponse{
			Message: fmt.Sprintf("[%d] Hello, %s!", i, req.GetName()),
		}); err != nil {
			return err
		}
		time.Sleep(time.Second * 1)
	}
	return nil
}

func (s *myServer) HelloClientStream(stream hellopb.GreetingService_HelloClientStreamServer) error {
	nameList := make([]string, 0)
	for {
		// クライアントストリームから明示的にリクエストを受け取る
		req, err := stream.Recv()
		// ストリームの終端をエラーの型から判別する
		if errors.Is(err, io.EOF) {
			message := fmt.Sprintf("Hello, %v!", nameList)
			// レスポンスを返してストリームをCloseしている？
			return stream.SendAndClose(&hellopb.HelloResponse{Message: message})
		}
		if err != nil {
			return err
		}
		nameList = append(nameList, req.GetName())
	}
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
