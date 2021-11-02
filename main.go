package main

import (
	"github.com/itnotebooks/fireye/action"
	"github.com/itnotebooks/fireye/config"
)

// 读取错误日志文件
func ReadErrorLog() {

}

func main() {

	// 解析配置文件中的信息放入到全局变量中
	conf := config.GetConfig()
	config.RenderConfig(conf)

	action.Start()

}
