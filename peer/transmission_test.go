package peer

import (
	"bufio"
	"testing"
	"time"
)

type TestIoReaderWriter struct {
	cache chan []byte
}

func (t *TestIoReaderWriter) Write(p []byte) (n int, err error) {
	t.cache <- p
	return len(p), nil
}

func (t *TestIoReaderWriter) Read(p []byte) (n int, err error) {
	p = <-t.cache
	return len(p), nil
}

func TestTransmission(t *testing.T) {
	trw := &TestIoReaderWriter{
		cache: make(chan []byte, 100),
	}

	rw := bufio.NewReadWriter(bufio.NewReader(trw), bufio.NewWriter(trw))

	trans := CreateTransmissionWithBufio(rw)

	go trans.RecvFile("peer.bk")
	go trans.SendFile("peer")

	time.Sleep(5 * time.Second)
}
