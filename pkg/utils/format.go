package utils

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/shopspring/decimal"
)

func ConvertToJsonString(v any) string {
	data, _ := json.Marshal(v)
	return string(data)
}

func ConvertToPercentage(numStr string) string {
	// 将字符串解析为 float64
	num, err := strconv.ParseFloat(numStr, 64)
	if err != nil {
		return "" // 解析失败返回空字符串
	}

	// 将小数乘以 100，转换为百分比数值
	percentage := num * 100
	// 格式化为字符串，保留两位小数
	return strconv.FormatFloat(percentage, 'f', 2, 64) + "%"
}

// GetDisplayWalletAddress 获取用于显示的用户名
func GetDisplayWalletAddress(walletAddress string) string {
	// 检查地址长度
	if len(walletAddress) > 9 {
		return fmt.Sprintf("%s...%s", walletAddress[:6], walletAddress[len(walletAddress)-4:])
	}
	// 如果地址不够长，直接返回原始地址
	return walletAddress
}

// FormatAmountWithDecimals 格式化金额
func FormatAmountWithDecimals(amount string, decimals int32) string {
	if amount == "" {
		return "0"
	}

	// 使用decimal包转换字符串为高精度数值
	amountDecimal, err := decimal.NewFromString(amount)
	if err != nil {
		return amount // 如果无法转换，直接返回原字符串
	}

	if amountDecimal.IsZero() {
		return "0"
	}

	// 使用高精度除法计算实际值
	actualAmount := amountDecimal.Shift(-decimals)

	// 转换为float进行后续格式化
	amountFloat, _ := actualAmount.Float64()

	// 如果金额很大，使用适当的格式
	if amountFloat >= 1000000 {
		return fmt.Sprintf("%.2fM", amountFloat/1000000)
	} else if amountFloat >= 1000 {
		return fmt.Sprintf("%.2fK", amountFloat/1000)
	}

	// 对于小数，保留合适的精度
	if amountFloat < 0.0001 {
		return actualAmount.Truncate(8).String()
	} else if amountFloat < 0.01 {
		return actualAmount.Truncate(6).String()
	} else if amountFloat < 1 {
		return actualAmount.Truncate(4).String()
	}

	return actualAmount.Truncate(2).String()
}

// FormatPrice 格式化价格
func FormatPrice(raw string) string {
	if raw == "" {
		return ""
	}

	// 转成浮点确保合法并去除科学计数等情况
	val, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return raw
	}
	if val == 0 {
		return "$0"
	}

	// 用足够精度转成字符串
	s := fmt.Sprintf("%.20f", val)        // 形如  "2.00003456000000000000"
	intPart, decPart := splitOnce(s, ".") // decPart 至少 1 位

	// 如果小数部分全是0，直接返回整数部分
	if strings.TrimRight(decPart, "0") == "" {
		return fmt.Sprintf("$%s", intPart)
	}

	// 1️⃣ 统计前导 0 个数
	zeroPrefix := 0
	for zeroPrefix < len(decPart) && decPart[zeroPrefix] == '0' {
		zeroPrefix++
	}

	// 2️⃣ 取首个非零数字开始的 4 位十进制数（含 0）
	start := zeroPrefix
	end := start + 4
	if end > len(decPart) {
		end = len(decPart)
	}
	digits := decPart[start:end]

	// 3️⃣ 拼接小数部分
	var frac string
	if zeroPrefix > 3 {
		frac = fmt.Sprintf("0{%d}%s", zeroPrefix, digits)
	} else {
		frac = strings.Repeat("0", zeroPrefix) + digits
	}

	return fmt.Sprintf("$%s.%s", intPart, frac)
}

// splitOnce 把 s 按第一个 sep 切成两段，若不存在 sep，则 decPart 为空串
func splitOnce(s, sep string) (intPart, decPart string) {
	if idx := strings.Index(s, sep); idx != -1 {
		return s[:idx], s[idx+1:]
	}
	return s, ""
}
