package sdk

import "net"

const (
	MsgTypeText = "text"
)

type Chat struct {
	Nick      string
	UserID    string
	SessionID string
	conn      *connect
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
	return &Chat{
		Nick:      nick,
		UserID:    userID,
		SessionID: sessionID,
		conn:      newConnet(ip, port),
	}
}

func (chat *Chat) Send(msg *Message) {
	chat.conn.send(msg)
}

// Close close chat
func (chat *Chat) Close() {
	chat.conn.close()
}

// Recv receive message
func (chat *Chat) Recv() <-chan *Message {
	return chat.conn.recv()
}
