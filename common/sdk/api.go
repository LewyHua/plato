package sdk

import (
	"encoding/json"
	"net"
	"sync"
	"time"

	"github.com/lewyhua/plato/common/idl/message"
	"github.com/lewyhua/plato/common/tcp"
	"google.golang.org/protobuf/proto"
)

const (
	MsgTypeText      = "text"
	MsgTypeAck       = "ack"
	MsgTypeReConn    = "reConn"
	MsgTypeHeartbeat = "heartbeat"
	MsgLogin         = "loginMsg"
)

type Chat struct {
	Nick             string
	UserID           string
	SessionID        string
	conn             *connect
	closeChan        chan struct{}
	MsgClientIDTable map[string]uint64 // sessionID -> clientID，记录每个客户端每个会话的客户端ID
	sync.RWMutex
}

type Message struct {
	Type       string
	Name       string
	FormUserID string
	ToUserID   string
	Content    string
	Session    string
}

// NewChat 传入网关IP和端口，创建一个新的Chat实例，包含用户昵称、用户ID和会话ID，以及conn
func NewChat(ip net.IP, port int, nick, userID, sessionID string) *Chat {

	chat := &Chat{
		Nick:             nick,
		UserID:           userID,
		SessionID:        sessionID,
		conn:             newConnet(ip, port),
		closeChan:        make(chan struct{}),
		MsgClientIDTable: make(map[string]uint64),
	}
	// 启动接收消息的协程
	go chat.loop()
	// 发送长链接创建消息
	chat.login()
	// 启动心跳协程
	go chat.heartbeat()
	return chat
}

// Send 客户端调用，发送上行消息
func (chat *Chat) Send(msg *Message) {
	data, _ := json.Marshal(msg)
	upMsg := &message.UPMsg{
		Head: &message.UPMsgHead{
			ClientID: chat.getClientID(msg.Session),
			ConnID:   chat.conn.connID,
		},
		UPMsgBody: data,
	}
	palyload, _ := proto.Marshal(upMsg)
	chat.conn.send(message.CmdType_UP, palyload)
}

// ReConn 客户端调用，重新连接
func (chat *Chat) ReConn() {
	chat.Lock()
	defer chat.Unlock()
	chat.conn.reConn() // 重新Dial进行连接，并绑定新的conn socket到conn结构体里面
	chat.reConn()      // 发送重连消息
}

// Close 客户端调用，关闭连接
func (chat *Chat) Close() {
	chat.conn.close()
	close(chat.closeChan)
	close(chat.conn.recvChan)
	close(chat.conn.sendChan)
}

// GetConnID 客户端调用，获取连接ID
// 在登录或重连时，服务端会返回一个connID，客户端可以通过这个方法获取当前连接的connID
func (chat *Chat) GetConnID() uint64 {
	return chat.conn.connID
}

// Recv 客户端调用，获取接收消息的通道
func (chat *Chat) Recv() <-chan *Message {
	return chat.conn.recv()
}

// loop 客户端调用，循环接收消息
func (chat *Chat) loop() {
Loop:
	for {
		select {
		case <-chat.closeChan:
			return
		default:
			mc := &message.MsgCmd{}
			data, err := tcp.ReadData(chat.conn.conn)
			if err != nil {
				goto Loop
			}
			err = proto.Unmarshal(data, mc)
			if err != nil {
				panic(err)
			}
			var msg *Message
			switch mc.Type {
			case message.CmdType_ACK: // 处理ACK消息
				msg = handAckMsg(chat.conn, mc.Payload)
			case message.CmdType_Push:
				msg = handPushMsg(chat.conn, mc.Payload)
			}
			chat.conn.recvChan <- msg
		}
	}
}

// getClientID 获取当前会话的客户端ID
func (chat *Chat) getClientID(sessionID string) uint64 {
	chat.Lock()
	defer chat.Unlock()
	var res uint64
	if id, ok := chat.MsgClientIDTable[sessionID]; ok {
		res = id
	}
	res++
	chat.MsgClientIDTable[sessionID] = res
	return res
}

// login 建立连接时发送登录消息
func (chat *Chat) login() {
	loginMsg := message.LoginMsg{
		Head: &message.LoginMsgHead{
			DeviceID: 123,
		},
	}
	payload, err := proto.Marshal(&loginMsg)
	if err != nil {
		panic(err)
	}
	chat.conn.send(message.CmdType_Login, payload)
}

// reConn 重新连接时发送重连消息
func (chat *Chat) reConn() {
	reConn := message.ReConnMsg{
		Head: &message.ReConnMsgHead{
			ConnID: chat.conn.connID,
		},
	}
	payload, err := proto.Marshal(&reConn)
	if err != nil {
		panic(err)
	}
	chat.conn.send(message.CmdType_ReConn, payload)
}

// heartbeat 定时发送心跳消息
func (chat *Chat) heartbeat() {
	tc := time.NewTicker(1 * time.Second)
	for {
		select {
		case <-chat.closeChan:
			return
		case <-tc.C:
			hearbeat := message.HeartbeatMsg{
				Head: &message.HeartbeatMsgHead{},
			}
			payload, err := proto.Marshal(&hearbeat)
			if err != nil {
				panic(err)
			}
			chat.conn.send(message.CmdType_Heartbeat, payload)
		}
	}
}
