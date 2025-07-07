package ipconf

import (
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/lewyhua/plato/common/config"
	"github.com/lewyhua/plato/ipconf/domain"
	"github.com/lewyhua/plato/ipconf/source"
)

func RunMain(path string) {
	config.Init(path) // 初始化配置
	source.Init()     // 初始化数据源
	domain.Init()     // 初始调度层
	s := server.Default(server.WithHostPorts(":6789"))
	s.GET("/ip/list", GetIPInfoList)
	s.Spin()
}
