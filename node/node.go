package node

import (
	"context"
	"fmt"
	"github.com/cmkqwerty/blocker/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
	"net"
	"sync"
)

type Node struct {
	version    string
	listenAddr string
	logger     *zap.SugaredLogger
	peerLock   sync.RWMutex
	peers      map[proto.NodeClient]*proto.Version
	proto.UnimplementedNodeServer
}

func NewNode() *Node {
	loggerConfig := zap.NewDevelopmentConfig()
	loggerConfig.EncoderConfig.TimeKey = ""
	logger, _ := loggerConfig.Build()
	return &Node{
		peers:   make(map[proto.NodeClient]*proto.Version),
		version: "0.0.1",
		logger:  logger.Sugar(),
	}
}

func (n *Node) Start(listenAddr string) error {
	n.listenAddr = listenAddr
	var (
		opts       []grpc.ServerOption
		grpcServer = grpc.NewServer(opts...)
	)
	ln, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return err
	}

	proto.RegisterNodeServer(grpcServer, n)

	n.logger.Infow("Starting node...", "on", listenAddr)
	return grpcServer.Serve(ln)
}

func (n *Node) BootstrapNetwork(bootstrapNodes []string) error {
	for _, node := range bootstrapNodes {
		client, err := makeNodeClient(node)
		if err != nil {
			return err
		}

		v, err := client.Handshake(context.TODO(), n.getVersion())
		if err != nil {
			n.logger.Error("handshake error")
			continue
		}

		n.addPeer(client, v)
	}

	return nil
}

func (n *Node) Handshake(ctx context.Context, v *proto.Version) (*proto.Version, error) {
	client, err := makeNodeClient(v.ListenAddr)
	if err != nil {
		return nil, err
	}

	n.addPeer(client, v)

	return n.getVersion(), nil
}

func (n *Node) HandleTransaction(ctx context.Context, tx *proto.Transaction) (*proto.Ack, error) {
	peerS, _ := peer.FromContext(ctx)
	fmt.Println("Received transaction from: ", peerS.Addr.String())
	return &proto.Ack{}, nil
}

func (n *Node) addPeer(client proto.NodeClient, v *proto.Version) {
	n.peerLock.Lock()
	defer n.peerLock.Unlock()

	n.logger.Debugw("New peer connected.", "addr", v.ListenAddr, "height", v.Height)

	n.peers[client] = v
}

func (n *Node) deletePeer(client proto.NodeClient) {
	n.peerLock.Lock()
	defer n.peerLock.Unlock()

	delete(n.peers, client)
}

func (n *Node) getVersion() *proto.Version {
	return &proto.Version{
		Version:    "0.0.1",
		Height:     0,
		ListenAddr: n.listenAddr,
	}
}

func makeNodeClient(listenAddr string) (proto.NodeClient, error) {
	conn, err := grpc.Dial(listenAddr, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	return proto.NewNodeClient(conn), nil
}
