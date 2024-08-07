package main

import (
	"ding/bot/difybot"
	dingbot "ding/bot/dingtalk"
	"ding/conf"
	"fmt"
)

func main() {

	err := conf.LoadConfig()
	if err != nil {
		fmt.Print("加载环境变量出错")
		return
	}

	// 初始化dify和钉钉机器人
	difybot.InitDifyClient()
	dingbot.StartDingRobot()

	//hertz http框架  未来支持提供接口调用
	//h := server.Default()
	//// 添加请求日志中间件
	//h.Use(middlewares.RequestLogger())
	//h.GET("/hello", handlers.TestTandlers.HelloHandler)
	//h.POST("/dify/chat-message", handlers.DifyTandlers.ChatMessageHandler)
	//h.Spin()
}
