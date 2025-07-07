package source

import (
	"fmt"

	"github.com/lewyhua/plato/common/discovery"
)

type EventType string

const (
	AddNodeEventType EventType = "add_node"
	DelNodeEventType EventType = "del_node"
)

var eventChan chan *Event

func EventChan() <-chan *Event {
	return eventChan
}

type Event struct {
	Type         EventType
	IP           string
	Port         string
	ConnectNum   float64
	MessageBytes float64
}

// NewEvent 创建一个新的事件对象
func NewEvent(ed *discovery.EndpointInfo) *Event {
	if ed == nil || ed.MetaData == nil {
		return nil
	}
	var connNum float64
	var msgBytes float64
	if data, ok := ed.MetaData["connect_num"]; ok {
		connNum = data.(float64) // 如果出错，此处应该panic 暴露错误
	}
	if data, ok := ed.MetaData["message_bytes"]; ok {
		msgBytes = data.(float64) // 如果出错，此处应该panic 暴露错误
	}
	return &Event{
		Type:         AddNodeEventType,
		IP:           ed.IP,
		Port:         ed.Port,
		ConnectNum:   connNum,
		MessageBytes: msgBytes,
	}
}
func (nd *Event) Key() string {
	return fmt.Sprintf("%s:%s", nd.IP, nd.Port)
}
