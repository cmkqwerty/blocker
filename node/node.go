package node

import (
	"context"
	"fmt"
	"github.com/cmkqwerty/blocker/proto"
	"google.golang.org/grpc/peer"
)

type Node struct {
	version string
	proto.UnimplementedNodeServer
}

func NewNode() *Node {
	return &Node{
		version: "0.0.1",
	}
}

func (n *Node) Handshake(ctx context.Context, v *proto.Version) (*proto.Version, error) {
	ourVersion := &proto.Version{
		Version: n.version,
		Height:  100,
	}

	peerS, _ := peer.FromContext(ctx)

	fmt.Printf("Received handshake from %s: %s\n", peerS.Addr.String(), v.Version)
	return ourVersion, nil
}

func (n *Node) HandleTransaction(ctx context.Context, tx *proto.Transaction) (*proto.Ack, error) {
	peerS, _ := peer.FromContext(ctx)
	fmt.Println("Received transaction from: ", peerS.Addr.String())
	return &proto.Ack{}, nil
}
