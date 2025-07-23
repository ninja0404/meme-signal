package kafka

import (
	"fmt"
	"os"
	"strings"
	"time"
)

func getClientID() string {
	// 获取主机名
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}

	// 清理主机名（移除特殊字符）
	hostname = strings.ReplaceAll(hostname, ".", "_")

	// 生成client ID: hostname-pid-timestamp
	clientID := fmt.Sprintf("%s-%d-%d",
		hostname,
		os.Getpid(),
		time.Now().Unix(),
	)

	return clientID
}
