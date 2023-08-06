package main

import (
	"bufio"
	"encoding/json"
	"time"

	"github.com/libp2p/go-libp2p/core/network"
)

type HeartBeat struct {
	Msg string `json:"msg"`
}

type SendFile struct {
	FileName string `json:"file_name"`
	Size     int    `json:"file_size"`
}

const (
	HEART_BEAT = 1
	SEND_FILE  = 2
)

type Request struct {
	MsgType   int        `json:"msg_type"`
	HeartBeat *HeartBeat `json:"heart_beat"`
	SendFile  *SendFile  `json:"send_file"`
}

type Response struct {
	MsgType int `json:"msg_type"`
	Code    int `json:"code"`
}

const (
	REQUEST  = 1
	RESPONSE = 2
)

type Msg struct {
	MsgType  int       `json:"msg_type"`
	Request  *Request  `json:"request"`
	Response *Response `json:"response"`
}

const (
	CLIENT = 1 // client start to send heart
	SERVER = 2 // server recv heart and response
)

type MsgDispatch struct {
	rw           *bufio.ReadWriter
	stream       network.Stream
	heartTime    int64
	side         int
	onServerFile func(name string, size int) bool
	onClientFile func(send bool)
}

func CreateMsgDispatch(stream network.Stream, side int, onServerFile func(name string, size int) bool, onClientFile func(send bool)) *MsgDispatch {
	return &MsgDispatch{
		rw:           bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream)),
		stream:       stream,
		side:         side,
		onServerFile: onServerFile,
		onClientFile: onClientFile,
	}
}

func CreateMsgDispatchWithBufio(rw *bufio.ReadWriter, side int, onServerFile func(name string, size int) bool, onClientFile func(send bool)) *MsgDispatch {
	return &MsgDispatch{
		rw:           rw,
		side:         side,
		onServerFile: onServerFile,
		onClientFile: onClientFile,
	}
}

func (m *MsgDispatch) Start() {
	go m.read()
	if m.side == CLIENT {
		go m.ClientHeartTimer()
	}
}

func (m *MsgDispatch) ConferSendFile(name string, size int) {
	msg := &Msg{
		MsgType: REQUEST,
		Request: &Request{
			MsgType: SEND_FILE,
			SendFile: &SendFile{
				FileName: name,
				Size:     size,
			},
		},
	}
	m.writeMsg(msg)
}

func (m *MsgDispatch) ClientHeartTimer() {
	ticker := time.NewTicker(15 * time.Second)
	for {
		<-ticker.C

		msg := &Msg{
			MsgType: REQUEST,
			Request: &Request{
				MsgType: HEART_BEAT,
				HeartBeat: &HeartBeat{
					Msg: "heart beat",
				},
			},
		}

		m.writeMsg(msg)
		log.Debugf("timer to send heartbeat.")
	}
}

func (m *MsgDispatch) read() {
	for {
		data := make([]byte, 1024)
		count, err := m.rw.Read(data)

		if err != nil {
			log.Errorf("read data from peer failed. err:%v", err)
			return
		}

		msg := Msg{}
		err = json.Unmarshal(data[:count], &msg)
		if err != nil {
			log.Errorf("read data from peer failed. err:%v", err)
			return
		}

		log.Debugf("get a msg. msg:%+v", msg)
		if msg.MsgType == REQUEST {
			req := msg.Request
			switch req.MsgType {
			case HEART_BEAT:
				m.onServerHeart(req)
			case SEND_FILE:
				m.onServerSendFile(req)
			}

		} else {
			resp := msg.Response
			switch resp.MsgType {
			case HEART_BEAT:
				m.onClientHeart(resp)
			case SEND_FILE:
				m.onClientSendFile(resp)
			}
		}
	}
}

func (m *MsgDispatch) onClientHeart(resp *Response) {
	m.heartTime = time.Now().Unix()
	log.Debugf("get a heartbeat response.")
}

func (m *MsgDispatch) onClientSendFile(resp *Response) {
	if resp.Code == 0 {
		m.onClientFile(true)
	} else {
		m.onClientFile(false)
	}
	log.Infof("get a send file response. code:%v", resp.Code)
}

func (m *MsgDispatch) onServerHeart(*Request) {
	log.Debugf("get a heartbeat request.")

	msg := &Msg{
		MsgType: RESPONSE,
		Response: &Response{
			MsgType: HEART_BEAT,
			Code:    0,
		},
	}
	m.writeMsg(msg)
}

func (m *MsgDispatch) onServerSendFile(req *Request) {
	recv := m.onServerFile(req.SendFile.FileName, req.SendFile.Size)
	msg := &Msg{
		MsgType: RESPONSE,
		Response: &Response{
			MsgType: SEND_FILE,
		},
	}
	if recv {
		msg.Response.Code = 0
	} else {
		msg.Response.Code = -1
	}
	m.writeMsg(msg)

	log.Infof("get a file request. name:%s, result:%v", req.SendFile.FileName, recv)
}
func (m *MsgDispatch) writeMsg(msg *Msg) error {
	sendBuf, err := json.Marshal(msg)
	if err != nil {
		log.Errorf("marshal response failed. err:%v", err)
		return err
	}
	return m.write(sendBuf)
}

func (m *MsgDispatch) write(data []byte) error {
	writeCount := 0
	for {
		count, err := m.rw.Write(data[writeCount:])
		if err != nil {
			log.Errorf("write data failed. err:%v", err)
			return err
		}
		writeCount += count
		if writeCount >= len(data) {
			m.rw.Flush()
			break
		}
	}
	return nil
}
