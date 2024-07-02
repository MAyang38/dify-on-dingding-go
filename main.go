package main

import (
	"ding/bot"
	"ding/clients"
	"ding/conf"
	"ding/consts"
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
	PermissionControlInit := os.Getenv("Permission_Control_Init")
	if PermissionControlInit == consts.PermissionNeedInit {
		err = clients.PermissionInit()
		if err != nil {
			fmt.Print("权限控制初始化失败")
			return
		}
	}
	// 初始化dify和钉钉机器人
	bot.InitDifyClient()
	bot.StartDingRobot()
}
