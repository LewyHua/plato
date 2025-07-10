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

func CancelConn(ctx *context.Context, endpoint string, connID uint64, payload []byte) error {
	rpcCtx, _ := context.WithTimeout(*ctx, 100*time.Millisecond)
	stateClient.CancelConn(rpcCtx, &service.StateRequest{
		Endpoint: endpoint,
		ConnID:   connID,
		Data:     payload,
	})
	return nil
}

func SendMsg(ctx *context.Context, endpoint string, connID uint64, payload []byte) error {
	log.Printf("SendMsg endpoint:%s, connID:%d, data:%s", endpoint, connID, string(payload))
	rpcCtx, _ := context.WithTimeout(*ctx, 100*time.Millisecond)
	_, err := stateClient.SendMsg(rpcCtx, &service.StateRequest{
		Endpoint: endpoint,
		ConnID:   connID,
		Data:     payload,
	})
	if err != nil {
		panic(err)
	}
	return nil
}
