package main

import (
	"ding/bot/difybot"
	dingbot "ding/bot/dingtalk"
	"ding/clients"
	"ding/conf"
	myconst "ding/consts"
	"fmt"
	"os"
)

func main() {

	err := conf.LoadConfig()
	if err != nil {
		fmt.Print("加载环境变量出错")
		return
	}
	// 机器人的权限控制 默认关闭
	privateServiceControlFlag := os.Getenv("Private_Service_Control_Flag")
	if privateServiceControlFlag == myconst.PrivateServiceNeedInit {
		err = clients.PermissionInit()
		if err != nil {
			fmt.Print("权限控制初始化失败")
			return
		}
		err = clients.QuestionLogInit()
		if err != nil {
			fmt.Print("机器人回答日志记录初始化失败")
			return
		}
	}
	// 初始化dify和钉钉机器人
	difybot.InitDifyClient()
	dingbot.StartDingRobot()

	// hertz http框架
	//h := server.Default()
	//// 添加请求日志中间件
	//h.Use(middlewares.RequestLogger())
	//h.GET("/hello", handlers.TestTandlers.HelloHandler)
	//h.POST("/dify/chat-message", handlers.DifyTandlers.ChatMessageHandler)
	//h.Spin()
}
