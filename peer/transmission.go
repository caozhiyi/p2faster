package peer

import (
	"bufio"
	"io"
	"os"

	"github.com/libp2p/go-libp2p/core/network"
)

type Transmission struct {
	rw     *bufio.ReadWriter
	stream network.Stream
}

func CreateTransmission(stream network.Stream) *Transmission {
	return &Transmission{
		rw:     bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream)),
		stream: stream,
	}
}

func CreateTransmissionWithBufio(rw *bufio.ReadWriter) *Transmission {
	return &Transmission{
		rw: rw,
	}
}

func (t *Transmission) RecvFile(path string) {
	f, err := os.Create(path)
	if err != nil {
		log.Errorf("create file failed. err:%v", err)
		return
	}
	defer f.Close()

	recvChan := make(chan []byte, 100)
	go t.read(recvChan)

	totalCount := 0
	for buf := range recvChan {
		count := 0
		totalCount += len(buf)
		log.Debugf("recv from chan:%d", totalCount)
		for {
			n, err := f.Write(buf[count:])
			if err != nil {
				log.Errorf("write data to file failed. err:%v", err)
			}
			count += n
			if count >= len(buf) {
				break
			}
		}
	}
	if t.stream != nil {
		t.stream.Close()
	}
}

func (t *Transmission) SendFile(path string) {
	file, err := os.Open(path)
	if err != nil {
		log.Errorf("open file failed. err:%v", err)
		return
	}
	defer file.Close()

	totalCount := 0
	buffer := make([]byte, 1024*10)
	done := false
	for {
		bytesread, err := file.Read(buffer[0:])
		if err != nil {
			if err != io.EOF {
				log.Errorf("read file failed. err:%v", err)
				return
			}
			done = true
		}

		totalCount += bytesread
		log.Debugf("send to stream:%d", totalCount)
		err = t.write(buffer[:bytesread])
		if err != nil {
			log.Errorf("write data failed. err:%v", err)
			done = true
		}

		if done {
			break
		}
	}
	if t.stream != nil {
		t.stream.Close()
	}
}

func (t *Transmission) read(out chan<- []byte) {
	defer close(out)
	done := false
	for {
		buf := make([]byte, 1024*10)
		num, err := t.rw.Read(buf)

		if err != nil {
			if err != io.EOF {
				log.Errorf("read data failed. err:%v", err)
				return
			}
			done = true
		}
		out <- buf[0:num]
		if done {
			break
		}
	}
}

func (t *Transmission) write(data []byte) error {
	writeCount := 0
	for {
		count, err := t.rw.Write(data[writeCount:])
		if err != nil {
			log.Errorf("write data failed. err:%v", err)
			return err
		}
		writeCount += count
		if writeCount >= len(data) {
			break
		}
	}
	t.rw.Flush()
	return nil
}
