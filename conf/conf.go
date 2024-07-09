package conf

import (
	"ding/consts"
	"fmt"
	"github.com/joho/godotenv"
	"os"
	"strings"
)

func LoadConfig() error {
	// 尝试加载 .env 文件
	err := godotenv.Load()
	if err != nil {
		// 如果 .env 文件不存在，尝试加载 .env_template 文件
		if os.IsNotExist(err) {
			err = godotenv.Load(".env_template")
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}
	// 从环境变量中获取关键词
	VoiceKeywords := os.Getenv("VOICE_KEYWORDS")
	if VoiceKeywords == "" {
		fmt.Println("No keywords found in environment")
	}
	// 将语音关键词字符串分割为 slice
	consts.VoicePrefix = strings.Split(VoiceKeywords, ",")
	return nil
}
