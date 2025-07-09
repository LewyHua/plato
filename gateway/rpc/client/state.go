// 创建state client，用来和 state server 进行通信

package client

import (
	"context"
	"log"
	"time"

	"github.com/lewyhua/plato/common/config"
	"github.com/lewyhua/plato/common/prpc"
	"github.com/lewyhua/plato/state/rpc/service"
)

var stateClient service.StateClient

func initStateClient() {
	pCli, err := prpc.NewPClient(config.GetStateServiceName())
	if err != nil {
		panic(err)
	}
	stateClient = service.NewStateClient(pCli.Conn())
}

func CancelConn(ctx *context.Context, endpoint string, fd int32, playLoad []byte) error {
	rpcCtx, _ := context.WithTimeout(*ctx, 100*time.Millisecond)
	stateClient.CancelConn(rpcCtx, &service.StateRequest{
		Endpoint: endpoint,
		Fd:       fd,
		Data:     playLoad,
	})
	return nil
}

func SendMsg(ctx *context.Context, endpoint string, fd int32, playLoad []byte) error {
	rpcCtx, _ := context.WithTimeout(*ctx, 100*time.Millisecond)
	log.Printf("STATE CLIENT: %v", stateClient)
	log.Printf("SendMsg endpoint:%s, fd:%d, data:%s", endpoint, fd, string(playLoad))
	_, err := stateClient.SendMsg(rpcCtx, &service.StateRequest{
		Endpoint: endpoint,
		Fd:       fd,
		Data:     playLoad,
	})
	if err != nil {
		panic(err)
	}
	return nil
}
