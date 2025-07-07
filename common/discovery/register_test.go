package discovery

import (
	"context"
	"log"
	"testing"
	"time"

	"github.com/lewyhua/plato/common/config"
)

func TestServiceRegister(t *testing.T) {
	// endpoints := []string{"localhost:2379"}
	// endpoints := []string{"http://my-etcd:2379"}
	config.Init("../../plato.yaml") // 初始化配置

	ctx := context.Background()

	// 注册服务
	ser, err := NewServiceRegister(&ctx, "/web/node1", &EndpointInfo{
		IP:   "127.0.0.1",
		Port: "9999",
	}, 5)
	if err != nil {
		t.Fatalf("Service register failed: %v", err)
	}
	defer ser.Close()

	log.Println("✅ Service registered successfully")

	// 开始监听续租响应
	go ser.ListenLeaseRespChan()

	time.Sleep(20 * time.Second)
	ser.Close()

	log.Println("🛑 Stopping service registration test")
}
