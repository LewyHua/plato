package sdk

import (
	"encoding/json"
	"fmt"
	"net"
	"sync/atomic"

	"github.com/lewyhua/plato/common/idl/message"
	"github.com/lewyhua/plato/common/tcp"
	"google.golang.org/protobuf/proto"
)

type connect struct {
	sendChan, recvChan chan *Message
	conn               *net.TCPConn // 本客户端和服务端的连接
	connID             uint64
	ip                 net.IP
	port               int
}

func newConnet(ip net.IP, port int) *connect {
	clientConn := &connect{
		sendChan: make(chan *Message),
		recvChan: make(chan *Message),
		ip:       ip,
		port:     port,
	}

	addr := &net.TCPAddr{IP: ip, Port: port}
	conn, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		panic(err)
	}
	clientConn.conn = conn
	return clientConn
}

// send 发送消息到网关
func (c *connect) send(ty message.CmdType, palyload []byte) {
	msgCmd := message.MsgCmd{
		Type:    ty,
		Payload: palyload,
	}
	msg, err := proto.Marshal(&msgCmd)
	if err != nil {
		panic(err)
	}
	dataPgk := tcp.DataPgk{
		Data: msg,
		Len:  uint32(len(msg)),
	}
	c.conn.Write(dataPgk.Marshal())

}

// 处理接收到的ack消息，如果是登录或重连消息，则获取/更新connID
func handAckMsg(c *connect, data []byte) *Message {
	ackMsg := &message.ACKMsg{}
	proto.Unmarshal(data, ackMsg)
	switch ackMsg.Type {
	case message.CmdType_Login, message.CmdType_ReConn:
		atomic.StoreUint64(&c.connID, ackMsg.ConnID)
	}
	return &Message{
		Type:       MsgTypeAck,
		Name:       "plato",
		FormUserID: "1212121",
		ToUserID:   "222212122",
		Content:    ackMsg.Msg,
	}
}

// handPushMsg 处理到网关的下行消息，回复ack消息，并返回消息内容
func handPushMsg(c *connect, data []byte) *Message {
	pushMsg := &message.PushMsg{}
	proto.Unmarshal(data, pushMsg)
	// if pushMsg.MsgID == c.maxMsgID+1 {
	// 	c.maxMsgID++
	msg := &Message{}
	json.Unmarshal(pushMsg.Content, msg)
	ackMsg := &message.ACKMsg{
		Type:   message.CmdType_UP,
		ConnID: c.connID,
	}
	ackData, _ := proto.Marshal(ackMsg)
	c.send(message.CmdType_ACK, ackData)
	return msg
	// }
}

// reConn 重连
func (c *connect) reConn() {
	c.conn.Close()
	addr := &net.TCPAddr{IP: c.ip, Port: c.port}
	conn, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		fmt.Printf("DialTCP.err=%+v", err)
	}
	c.conn = conn
}

func (c *connect) recv() <-chan *Message {
	return c.recvChan
}

func (c *connect) close() {
	// 目前没啥值得回收的
	c.conn.Close()
}
