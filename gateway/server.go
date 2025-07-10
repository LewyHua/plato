package gateway

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"

	"github.com/lewyhua/plato/common/config"
	"github.com/lewyhua/plato/common/prpc"
	"github.com/lewyhua/plato/common/tcp"
	"github.com/lewyhua/plato/gateway/rpc/client"
	"github.com/lewyhua/plato/gateway/rpc/service"
	"google.golang.org/grpc"
)

var cmdChannel chan *service.CmdContext

// RunMain 启动网关服务
func RunMain(path string) {
	config.Init(path)
	ln, err := net.ListenTCP("tcp", &net.TCPAddr{Port: config.GetGatewayTCPServerPort()})
	if err != nil {
		log.Fatalf("StartTCPEPollServer err:%s", err.Error())
		panic(err)
	}
	initWorkPoll() // 初始化协程池，1024个worker
	initEpoll(ln, runProc)
	fmt.Println("-------------IM gateway stated------------")

	cmdChannel = make(chan *service.CmdContext, config.GetGatewayCmdChannelNum())

	// 创建prpc server
	s := prpc.NewPServer(
		prpc.WithServiceName(config.GetGatewayServiceName()),
		prpc.WithIP(config.GetGatewayServiceAddr()),
		prpc.WithPort(config.GetGatewayRPCServerPort()), prpc.WithWeight(config.GetGatewayRPCWeight()))

	fmt.Println(config.GetGatewayServiceName(), config.GetGatewayServiceAddr(), config.GetGatewayRPCServerPort(), config.GetGatewayRPCWeight())

	// 注册网关服务到prpc server
	s.RegisterService(func(server *grpc.Server) {
		service.RegisterGatewayServer(server, &service.Service{CmdChannel: cmdChannel})
	})

	// 启动 state rpc 客户端
	client.Init()
	// 启动 命令处理写协程
	go cmdHandler()
	// 启动 rpc server
	s.Start(context.TODO())
}

// epoll wait 的回调处理函数
func runProc(c *connection, ep *epoller) {
	ctx := context.Background() // 起始的contenxt
	// step1: 以LT模式读取，所以一次只读取一个datā包
	dataBuf, err := tcp.ReadData(c.conn)
	if err != nil {
		// 如果读取conn时发现连接关闭，则直接端口连接
		// 通知 state 清理掉意外退出的 conn的状态信息
		if errors.Is(err, io.EOF) {
			// 这步操作是异步的，不需要等到返回成功在进行，因为消息可靠性的保障是通过协议完成的而非某次cmd
			ep.remove(c) // 从红黑树中删除该连接，以及从ep.tables中删除该映射
			client.CancelConn(&ctx, getEndpoint(), c.id, nil)
		}
		return
	}
	err = wPool.Submit(func() {
		// step2:交给 state server rpc 处理
		client.SendMsg(&ctx, getEndpoint(), c.id, dataBuf)
	})
	if err != nil {
		log.Printf("runProc:err:%+v\n", err.Error())
	}
}

// 处理两种命令：删除连接、和发送消息到连接
func cmdHandler() {
	for cmd := range cmdChannel {
		// 异步提交到协池中完成发送任务
		switch cmd.Cmd {
		case service.DelConnCmd:
			wPool.Submit(func() { closeConn(cmd) })
		case service.PushCmd:
			wPool.Submit(func() { sendMsgByCmd(cmd) })
		default:
			panic("command undefined")
		}
	}
}

// 关闭连接的处理函数
func closeConn(cmd *service.CmdContext) {
	log.Println("close:", cmd.ConnID)
	subTcpNum()
	if connPtr, ok := ep.tables.Load(cmd.ConnID); ok {
		conn, _ := connPtr.(*connection)
		conn.Close() // 因为只有gateway有这个fd，所以关闭时会自动从epoll中删除
		ep.tables.Delete(cmd.ConnID)
		conn.e.fdToConnTable.Delete(conn.fd) // 从epoll的fd映射表中删除
	}
}

// 发送消息到连接的处理函数
func sendMsgByCmd(cmd *service.CmdContext) {
	if connPtr, ok := ep.tables.Load(cmd.ConnID); ok {
		conn, _ := connPtr.(*connection)
		dp := tcp.DataPgk{
			Len:  uint32(len(cmd.Payload)),
			Data: cmd.Payload,
		}
		tcp.SendData(conn.conn, dp.Marshal())
	}
}

func getEndpoint() string {
	return fmt.Sprintf("%s:%d", config.GetGatewayServiceAddr(), config.GetGatewayRPCServerPort())
}
