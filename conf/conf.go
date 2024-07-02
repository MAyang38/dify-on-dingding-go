package conf

import (
	"github.com/joho/godotenv"
	"os"
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
	return nil
}
