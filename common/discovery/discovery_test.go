package discovery

import (
	"context"
	"log"
	"testing"
	"time"

	"github.com/lewyhua/plato/common/config"
)

func TestServiceDiscovery(t *testing.T) {
	// endpoints := []string{"localhost:2379"}
	// endpoints := []string{"http://my-etcd:2379"}
	config.Init("../../plato.yaml") // 初始化配置

	ctx := context.Background()

	ser := NewServiceDiscovery(&ctx)
	defer ser.Close()

	// 注册监听回调
	ser.WatchService("/web/",
		func(key, value string) {
			log.Printf("🟢 [WEB] Service added: %s => %s\n", key, value)
		},
		func(key, value string) {
			log.Printf("🔴 [WEB] Service removed: %s => %s\n", key, value)
		})

	ser.WatchService("/gRPC/",
		func(key, value string) {
			log.Printf("🟢 [gRPC] Service added: %s => %s\n", key, value)
		},
		func(key, value string) {
			log.Printf("🔴 [gRPC] Service removed: %s => %s\n", key, value)
		})

	// 监听 30 秒
	log.Println("🔍 Listening for service changes...")
	time.Sleep(30 * time.Second)
	log.Println("🛑 Stopping service discovery test")
}
