package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"

	hellopb "Practice-gRPC/pkg/grpc"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	scanner *bufio.Scanner
	client  hellopb.GreetingServiceClient
)

func main() {
	fmt.Println("start gRPC Client.")

	// 標準入力を受け取る受け取る受け取るスキャナを用意
	scanner = bufio.NewScanner(os.Stdin)

	// gRPCサーバとのコネクションの確立
	address := "localhost:8080"
	conn, err := grpc.Dial(
		address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		log.Fatal("Connection failed.")
		return
	}
	defer conn.Close()

	// gPRCクライアントの生成
	client = hellopb.NewGreetingServiceClient(conn)

	for {
		fmt.Println("1: send Request")
		fmt.Println("2: HelloServerStream")
		fmt.Println("3: HelloClientStream")
		fmt.Println("4: exit")
		fmt.Print("please enter >")

		scanner.Scan()
		input := scanner.Text()

		switch input {
		case "1":
			Hello()
		case "2":
			HelloServerStream()
		case "3":
			HelloClientStream()
		case "4":
			fmt.Println("bye.")
			goto M
		}
	}
M:
}

func Hello() {
	fmt.Println("Please enter your name.")
	scanner.Scan()
	name := scanner.Text()

	req := &hellopb.HelloRequest{Name: name}
	res, err := client.Hello(context.Background(), req)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(res.GetMessage())
	}
}

func HelloServerStream() {
	fmt.Println("Please enter your name.")
	scanner.Scan()
	name := scanner.Text()

	req := &hellopb.HelloRequest{Name: name}
	// サーバーから複数回レスポンスを受け取るためのストリームを得る
	stream, err := client.HelloServerStream(context.Background(), req)
	if err != nil {
		fmt.Println(err)
		return
	}

	for {
		// ストリームからレスポンスを受け取る
		res, err := stream.Recv()
		// ストリームの終端（全てのレスポンスを返したか）を確認する
		if errors.Is(err, io.EOF) {
			fmt.Println("all the responses have already received.")
			break
		}

		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(res)
	}
}

func HelloClientStream() {
	// サーバーにリクエストを送るためのストリームを生成
	stream, err := client.HelloClientStream(context.Background())
	if err != nil {
		fmt.Println(err)
		return
	}

	sendCount := 5
	fmt.Printf("Please enter %d names.\n", sendCount)
	for i := 0; i < sendCount; i++ {
		scanner.Scan()
		name := scanner.Text()

		// ストリームからリクエストを送信
		if err := stream.Send(&hellopb.HelloRequest{Name: name}); err != nil {
			fmt.Println(err)
			return
		}
	}

	// ストリームをCloseしてレスポンスを受け取る
	res, err := stream.CloseAndRecv()
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(res.GetMessage())
	}
}
