package peer

import (
	"bufio"
	"fmt"
	"os"

	"github.com/libp2p/go-libp2p/core/network"
)

type Chat struct {
	rw     *bufio.ReadWriter
	stream network.Stream
}

func CreateChat(stream network.Stream) *Chat {
	return &Chat{
		rw:     bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream)),
		stream: stream,
	}
}

func CreateChatWithBufio(rw *bufio.ReadWriter) *Chat {
	return &Chat{
		rw: rw,
	}
}

func (c *Chat) Start() {
	go c.read()
	go c.write()
}

func (t *Chat) read() {
	for {
		str, _ := t.rw.ReadString('\n')

		if str == "" {
			return
		}
		if str != "\n" {
			// Green console colour: 	\x1b[32m
			// Reset console colour: 	\x1b[0m
			fmt.Printf("\x1b[32m%s\x1b[0m> ", str)
		}

	}
}

func (t *Chat) write() {
	stdReader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("> ")
		sendData, err := stdReader.ReadString('\n')
		if err != nil {
			log.Error(err)
			return
		}

		t.rw.WriteString(fmt.Sprintf("%s\n", sendData))
		t.rw.Flush()
	}
}
