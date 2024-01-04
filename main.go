package main

import (
	"context"
	"fmt"
	"github.com/cmkqwerty/blocker/node"
	"github.com/cmkqwerty/blocker/proto"
	"google.golang.org/grpc"
	"log"
	"net"
	"time"
)

func main() {
	node := node.NewNode()

	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)

	ln, err := net.Listen("tcp", ":3000")
	if err != nil {
		log.Fatal(err)
	}

	proto.RegisterNodeServer(grpcServer, node)
	fmt.Println("Node running on port: 3000")

	go func() {
		for {
			time.Sleep(3 * time.Second)
			makeTransaction()
		}
	}()

	grpcServer.Serve(ln)
}

func makeTransaction() {
	client, err := grpc.Dial("localhost:3000", grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}

	c := proto.NewNodeClient(client)

	tx := &proto.Version{
		Version: "0.0.1",
		Height:  1,
	}

	_, err = c.Handshake(context.Background(), tx)
	if err != nil {
		log.Fatal(err)
	}
}
