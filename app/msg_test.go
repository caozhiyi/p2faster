package main

import (
	"bufio"
	"testing"
	"time"

	logging "github.com/ipfs/go-log/v2"
)

type TestIoReaderWriter struct {
	cache chan []byte
}

func (t *TestIoReaderWriter) Write(p []byte) (n int, err error) {
	t.cache <- p
	return len(p), nil
}

func (t *TestIoReaderWriter) Read(p []byte) (n int, err error) {
	buf := <-t.cache
	copy(p, buf)
	return len(buf), nil
}

func TestMsgDispatch(t *testing.T) {
	logging.SetLogLevel("ui", "debug")
	trw := &TestIoReaderWriter{
		cache: make(chan []byte, 100),
	}

	rw := bufio.NewReadWriter(bufio.NewReader(trw), bufio.NewWriter(trw))

	serverDispatcher := CreateMsgDispatchWithBufio(
		rw,
		SERVER,
		func(name string, size int) bool {
			log.Infof("server get a send file request. name:%v, size:%v", name, size)
			return true
		},
		func(recv bool) {
			log.Infof("server get a send file respnse. recv:%v", recv)
		},
	)
	serverDispatcher.Start()

	clientDispatcher := CreateMsgDispatchWithBufio(
		rw,
		CLIENT,
		func(name string, size int) bool {
			log.Infof("client get a send file request. name:%v, size:%v", name, size)
			return true
		},
		func(recv bool) {
			log.Infof("client get a send file respnse. recv:%v", recv)
		},
	)
	clientDispatcher.Start()

	clientDispatcher.ConferSendFile("client file name", 10234)
	serverDispatcher.ConferSendFile("server file name", 10234)

	time.Sleep(60 * time.Second)
}
