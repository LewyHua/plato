package sdk

import (
	"encoding/json"
	"net"

	"github.com/lewyhua/plato/common/tcp"
)

type connect struct {
	sendChan, recvChan chan *Message
	conn               *net.TCPConn // 本客户端和服务端的连接
}

func newConnet(ip net.IP, port int) *connect {
	clientConn := &connect{
		sendChan: make(chan *Message),
		recvChan: make(chan *Message),
	}

	addr := &net.TCPAddr{IP: ip, Port: port}
	conn, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		panic(err)
	}
	clientConn.conn = conn
	go func() {
		for {
			data, err := tcp.ReadData(conn)
			if err != nil {
				panic(err)
			}
			msg := &Message{}
			err = json.Unmarshal(data, msg)
			if err != nil {
				panic(err)
			}
			clientConn.recvChan <- msg
		}
	}()
	return clientConn
}

func (c *connect) send(data *Message) {
	bytes, _ := json.Marshal(data)
	dataPgk := tcp.DataPgk{
		Len:  uint32(len(bytes)),
		Data: bytes,
	}
	xx := dataPgk.Marshal()
	c.conn.Write(xx)
}

func (c *connect) recv() <-chan *Message {
	return c.recvChan
}

func (c *connect) close() {
	// 目前没啥值得回收的
	c.conn.Close()
}
