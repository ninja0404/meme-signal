package utils

import (
	"crypto/rand"
	"math/big"
)

// RandString 生成指定长度的随机字符串
func RandString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	charsetLength := big.NewInt(int64(len(charset)))

	for i := 0; i < length; i++ {
		n, err := rand.Int(rand.Reader, charsetLength)
		if err != nil {
			// 如果生成随机数失败，使用固定字符
			result[i] = charset[0]
			continue
		}
		result[i] = charset[n.Int64()]
	}

	return string(result)
}
