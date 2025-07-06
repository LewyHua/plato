package source

import "context"

func Init() {
	eventChan = make(chan *Event)
	ctx := context.Background()
	go DataHandler(&ctx)
}

// DataHandler 服务发现处理
func DataHandler(ctx *context.Context) {
	// 这里可以添加数据处理逻辑
	// 例如监听事件通道，处理数据等
	for event := range EventChan() {
		switch event.Type {
		case AddNodeEvent:
			// 处理添加节点事件
		case DelNodeEvent:
			// 处理删除节点事件
		default:
			// 处理其他类型的事件
		}
	}
}
