package peer

import (
	"context"
	"fmt"
	"time"

	logging "github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/libp2p/go-libp2p/p2p/protocol/circuitv2/client"
	ma "github.com/multiformats/go-multiaddr"
)

// TODO dynamic get
var relayId = "12D3KooW9qaj35NgxHjtrH6uKEKKgE1iPhYjrKpTKnDh1mUeGCAh"
var relayIp = "/ip4/172.31.52.90/tcp/7785"

const ChatProtocol protocol.ID = "/chatStream"
const FileSendProtocol protocol.ID = "/sendStream"

var log = logging.Logger("peer")

type BinaryConn struct {
	localNode    host.Host
	relayInfo    *peer.AddrInfo
	peerInfo     *peer.AddrInfo
	chatStream   network.Stream
	onFileStream func(network.Stream)
	onChatStream func(network.Stream)
	onCreate     func(string)
}

func CreateBinaryConn(onFileStream, onChatStream func(network.Stream), onCreate func(string)) *BinaryConn {
	return &BinaryConn{
		onFileStream: onFileStream,
		onChatStream: onChatStream,
		onCreate:     onCreate,
	}
}

func (c *BinaryConn) Init() error {
	err := c.localInit()
	if err != nil {
		return err
	}

	return nil
}

func (c *BinaryConn) Connect(peerId string) (network.Stream, error) {
	if len(peerId) > 0 {
		return c.connectPeer(peerId)
	}
	return nil, fmt.Errorf("invlied peer id")
}

func (c *BinaryConn) CreateSendStream() (network.Stream, error) {
	if c.chatStream == nil {
		return nil, fmt.Errorf("invalid connecton")
	}
	s, err := c.localNode.NewStream(network.WithUseTransient(context.Background(), "sendStream"), c.chatStream.Conn().RemotePeer(), FileSendProtocol)
	if err != nil {
		log.Errorf("Whoops, this should have worked...: ", err)
		return nil, err
	}

	return s, nil
}

func (c *BinaryConn) localInit() error {
	var err error
	c.localNode, err = libp2p.New(
		libp2p.EnableNATService(),
		libp2p.EnableRelayService(),
		libp2p.EnableRelay(),
		libp2p.EnableHolePunching(),
	)
	if err != nil {
		log.Infof("failed to create local host. err:%v", err)
		return err
	}
	log.Infof("listen addresses:", c.localNode.Addrs())

	c.relayInfo, err = connectRelay(c.localNode, relayId, relayIp)
	if err != nil {
		return err
	}

	c.localNode.SetStreamHandler(ChatProtocol, func(s network.Stream) {
		log.Infof("get a chat stream.")
		c.chatStream = s
		c.onChatStream(s)
	})
	c.localNode.SetStreamHandler(FileSendProtocol, func(s network.Stream) {
		log.Infof("get a send stream.")
		c.onFileStream(s)
	})

	c.onCreate(c.localNode.ID().String())
	return nil
}

func (c *BinaryConn) connectPeer(peerId string) (network.Stream, error) {
	if c.chatStream != nil {
		return nil, fmt.Errorf("already connected")
	}
	relayaddr, err := ma.NewMultiaddr("/p2p/" + c.relayInfo.ID.String() + "/p2p-circuit/p2p/" + peerId)
	if err != nil {
		log.Errorf("connect to peer failed. err:%v", err)
		return nil, err
	}

	c.peerInfo, _ = peer.AddrInfoFromP2pAddr(relayaddr)
	if err := c.localNode.Connect(context.Background(), *c.peerInfo); err != nil {
		log.Errorf("Unexpected error here. Failed to connect unreachable1 and unreachable2: %v", err)
		return nil, err
	}

	time.Sleep(time.Duration(5) * time.Second)

	s, err := c.localNode.NewStream(network.WithUseTransient(context.Background(), "chatStream"), c.peerInfo.ID, ChatProtocol)
	if err != nil {
		log.Errorf("Whoops, this should have worked...: ", err)
		return nil, err
	}
	c.chatStream = s
	return s, nil
}

func connectRelay(host host.Host, relayId, relayAddr string) (*peer.AddrInfo, error) {
	ms1, _ := ma.NewMultiaddr(fmt.Sprintf("%s/p2p/%v", relayAddr, relayId))
	relayInfo, _ := peer.AddrInfoFromP2pAddr(ms1)

	if err := host.Connect(context.Background(), *relayInfo); err != nil {
		log.Errorf("failed to connect relay server. addr:%s, err: %v", relayAddr, err)
		return nil, err
	}

	_, err := client.Reserve(context.Background(), host, *relayInfo)
	if err != nil {
		log.Errorf("failed to receive a relay. addr:%s, err:%v", relayAddr, err)
		return nil, err
	}

	return relayInfo, nil
}
