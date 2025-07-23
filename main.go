package main

import (
	"fmt"

	"github.com/ninja0404/meme-signal/internal/app"
)

func main() {
	// 创建应用实例
	application := app.New()

	// 启动应用
	if err := application.Start("./config/config.yaml"); err != nil {
		fmt.Printf("应用启动失败: %v\n", err)
		return
	}
}
