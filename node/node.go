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

func (n *Node) Start(listenAddr string, bootstrapNodes []string) error {
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

	// bootstrap network with the known nodes
	if len(bootstrapNodes) > 0 {
		go func() {
			err := n.bootstrapNetwork(bootstrapNodes)
			if err != nil {
				n.logger.Errorw("Bootstrap error", "error", err)
			}
		}()
	}
	return grpcServer.Serve(ln)
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

func (n *Node) bootstrapNetwork(bootstrapNodes []string) error {
	for _, node := range bootstrapNodes {
		if !n.canConnectWith(node) {
			continue
		}
		n.logger.Debugw("Dialing remote nodes...", "ourNode", n.listenAddr, "remoteNode", node)

		client, v, err := n.dialRemoteNode(node)
		if err != nil {
			return err
		}

		n.addPeer(client, v)
	}

	return nil
}

func (n *Node) addPeer(client proto.NodeClient, v *proto.Version) {
	n.peerLock.Lock()
	defer n.peerLock.Unlock()

	n.peers[client] = v

	if len(v.PeerList) > 0 {
		go func() {
			go func() {
				err := n.bootstrapNetwork(v.PeerList)
				if err != nil {
					n.logger.Errorw("Bootstrap error", "error", err)
				}
			}()
		}()
	}

	n.logger.Debugw("New peer successfully connected.",
		"ourNode", n.listenAddr,
		"remoteNode", v.ListenAddr,
		"height", v.Height)
}

func (n *Node) deletePeer(client proto.NodeClient) {
	n.peerLock.Lock()
	defer n.peerLock.Unlock()

	delete(n.peers, client)
}

func (n *Node) dialRemoteNode(addr string) (proto.NodeClient, *proto.Version, error) {
	client, err := makeNodeClient(addr)
	if err != nil {
		return nil, nil, err
	}

	v, err := client.Handshake(context.Background(), n.getVersion())
	if err != nil {
		return nil, nil, err
	}

	return client, v, nil
}

func (n *Node) getVersion() *proto.Version {
	return &proto.Version{
		Version:    "0.0.1",
		Height:     0,
		ListenAddr: n.listenAddr,
		PeerList:   n.getPeerList(),
	}
}

func (n *Node) canConnectWith(addr string) bool {
	if n.listenAddr == addr {
		return false
	}

	connectedPeers := n.getPeerList()
	for _, connectedAddr := range connectedPeers {
		if addr == connectedAddr {
			return false
		}
	}

	return true
}

func (n *Node) getPeerList() []string {
	n.peerLock.RLock()
	defer n.peerLock.RUnlock()

	var peers []string
	for _, v := range n.peers {
		peers = append(peers, v.ListenAddr)
	}

	return peers
}

func makeNodeClient(listenAddr string) (proto.NodeClient, error) {
	conn, err := grpc.Dial(listenAddr, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	return proto.NewNodeClient(conn), nil
}
