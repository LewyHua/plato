package source

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/lewyhua/plato/common/config"
	"github.com/lewyhua/plato/common/discovery"
)

// func init() {
// 	ctx := context.Background()
// 	testServiceRegister(&ctx, "7896", "node1")
// 	testServiceRegister(&ctx, "7897", "node2")
// 	testServiceRegister(&ctx, "7898", "node3")
// }

func testServiceRegister(ctx *context.Context, port, node string) {
	// 模拟服务发现
	log.Printf("🔍 Starting service registration for node %s on port %s", node, port)
	go func() {
		ed := discovery.EndpointInfo{
			IP:   "127.0.0.1",
			Port: port,
			MetaData: map[string]interface{}{
				"connect_num":   float64(rand.Int63n(12312321231231131)),
				"message_bytes": float64(rand.Int63n(1231232131556)),
			},
		}
		log.Printf("🔍 Initializing service discovery for node %s with endpoint: %+v", node, ed)
		// 创建服务注册对象
		sr, err := discovery.NewServiceRegister(ctx, fmt.Sprintf("%s/%s", config.GetServicePathForIPConf(), node), &ed, time.Now().Unix())
		if err != nil {
			log.Printf("❌ Failed to create service register for node %s: %v", node, err)
			panic(err)
		}
		log.Printf("✅ Service register created successfully for node %s", node)
		go sr.ListenLeaseRespChan()
		for {
			ed = discovery.EndpointInfo{
				IP:   "127.0.0.1",
				Port: port,
				MetaData: map[string]interface{}{
					"connect_num":   float64(rand.Int63n(12312321231231131)),
					"message_bytes": float64(rand.Int63n(1231232131556)),
				},
			}
			sr.UpdateValue(&ed)
			time.Sleep(1 * time.Second)
		}
	}()
}
