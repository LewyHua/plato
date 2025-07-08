package gateway

import (
	"fmt"
	"log"
	"net"

	"github.com/lewyhua/plato/common/config"
	"github.com/lewyhua/plato/common/tcp"
)

// RunMain 启动网关服务
func RunMain(path string) {
	config.Init(path)
	log.Println("Listening on port:", config.GetGatewayServerPort())
	listener, err := net.ListenTCP("tcp", &net.TCPAddr{Port: config.GetGatewayServerPort()})
	if err != nil {
		log.Fatalf("StartTCPEPollServer err:%s", err.Error())
		panic(err)
	}
	initWorkPoll() // 初始化工作池
	initEpoll(listener, runProc)
	fmt.Println("-------------im gateway stated------------")
	select {}
}

func runProc(conn *net.TCPConn, ep *epoller) {
	// step1: 读取一个完整的消息包
	dataBuf, err := tcp.ReadData(conn)
	if err != nil {
		return
	}
	err = wPool.Submit(func() {
		// step2:交给 state server rpc 处理
		bytes := tcp.DataPgk{
			Len:  uint32(len(dataBuf)),
			Data: dataBuf,
		}
		// 目前只是直接返回给了客户端，后续可以改为调用 state server rpc
		tcp.SendData(conn, bytes.Marshal())
	})
	if err != nil {
		log.Printf("runProc:err:%+v\n", err.Error())
	}
}
