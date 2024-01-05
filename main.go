package main

import (
	"context"
	"github.com/cmkqwerty/blocker/node"
	"github.com/cmkqwerty/blocker/proto"
	"google.golang.org/grpc"
	"log"
)

func main() {
	makeNode("localhost:3000", []string{})
	makeNode("localhost:3001", []string{"localhost:3000"})

	select {}
}

func makeNode(listenAddr string, bootstrapNodes []string) *node.Node {
	n := node.NewNode()
	go func() {
		log.Fatal(n.Start(listenAddr))
	}()
	if len(bootstrapNodes) > 0 {
		if err := n.BootstrapNetwork(bootstrapNodes); err != nil {
			log.Fatal(err)
		}
	}

	return n
}

func makeTransaction() {
	client, err := grpc.Dial("localhost:3000", grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}

	c := proto.NewNodeClient(client)

	version := &proto.Version{
		Version:    "0.0.1",
		Height:     1,
		ListenAddr: "localhost:3001",
	}

	_, err = c.Handshake(context.TODO(), version)
	if err != nil {
		log.Fatal(err)
	}
}
