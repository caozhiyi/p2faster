//main.go

package peer

import (
	"flag"
	"p2faster/peer"

	logging "github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p/core/network"
)

var log = logging.Logger("test")

func main() {
	dist := flag.String("d", "", "your fanal nodeID")
	flag.Parse()

	logging.SetLogLevel("p2p-holepunch", "debug")
	logging.SetLogLevel("peer", "info")
	//*dist = "12D3KooWSZoaavzgnJ4dNJaraJspT1Kp6M3CAvJUfjdjXiDr6Uca"

	conn := peer.CreateBinaryConn(
		func(s network.Stream) {
			chat := peer.CreateChat(s)
			chat.Start()
		}, func(s network.Stream) {
			chat := peer.CreateChat(s)
			chat.Start()
		}, onId)

	conn.Init()
	conn.Connect(*dist)

	select {}
}

func onId(id string) {
	log.Infof("local node id:%s", id)
}
