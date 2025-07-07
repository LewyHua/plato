package config

import (
	"time"

	"github.com/bytedance/gopkg/util/logger"
	"github.com/spf13/viper"
)

func Init(path string) {
	viper.SetConfigFile(path)
	viper.SetConfigType("yaml")
	logger.Infof("Using config file: %s", path)
	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}
}

// 获取服务发现的地址
func GetEndpointsForDiscovery() []string {
	return viper.GetStringSlice("discovery.endpoints")
}

// 获取连接服务发现集群的超时时间 单位：秒
func GetTimeoutForDiscovery() time.Duration {
	return viper.GetDuration("discovery.timeout") * time.Second
}

// 获取服务发现的根路径
func GetServicePathForIPConf() string {
	return viper.GetString("ip_conf.service_path")
}

// 判断是不是debug环境
func IsDebug() bool {
	env := viper.GetString("global.env")
	return env == "debug"
}
