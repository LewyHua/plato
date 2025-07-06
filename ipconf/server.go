package ipconf

import "github.com/cloudwego/hertz/pkg/app/server"

func RunMain(path string) {
	config.Init(path)
	source.Init() // 初始化数据源
	domain.Init() // 初始调度层
	s := server.Default(server.WithHostPorts(":6789"))
	s.GET("/ip/list", GetIPInfoList)
	s.Spin()
}
