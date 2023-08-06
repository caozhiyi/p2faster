// relayserver.go
package main

import (
	"log"

	logging "github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/p2p/protocol/circuitv2/relay"
)

func main() {
	run()
}

func run() {

	logging.SetLogLevel("p2p-holepunch", "debug")
	logging.SetLogLevel("peer", "info")
	logging.SetLogLevel("relay", "debug")
	opts := []libp2p.Option{
		libp2p.ListenAddrStrings(
			"/ip4/0.0.0.0/tcp/7785",
			"/ip4/0.0.0.0/udp/7786/quic",
			"/ip4/0.0.0.0/udp/7786/quic-v1",
			"/ip6/::0/tcp/5021",
			"/ip6/::0/udp/5022/quic",
			"/ip6/::0/udp/5022/quic-v1",
		),
		libp2p.EnableNATService(),
		libp2p.EnableRelayService(),
		libp2p.EnableRelay(),
		libp2p.EnableHolePunching(),
	}
	host, err := libp2p.New(opts...)
	if err != nil {
		log.Printf("Failed to create relay1: %v", err)
		return
	}

	_, err = relay.New(host)
	if err != nil {
		log.Printf("Failed to instantiate the relay: %v", err)
		return
	}

	log.Printf("relay1Info ID: %v Addrs: %v", host.ID(), host.Addrs())

	select {}
}
